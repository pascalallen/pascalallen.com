package feed

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/feed"
)

const (
	CacheTTL      = 15 * time.Minute
	retryAfterErr = time.Minute
)

type Clock func() time.Time

type entry struct {
	items   []feed.Item
	expires time.Time
}

type call struct {
	done  chan struct{}
	items []feed.Item
	err   error
}

type Cache struct {
	now    Clock
	ttl    time.Duration
	logger *slog.Logger

	mu       sync.Mutex
	entries  map[string]entry
	inflight map[string]*call
}

func NewCache(logger *slog.Logger) *Cache {
	return NewCacheWithClock(logger, time.Now, CacheTTL)
}

func NewCacheWithClock(logger *slog.Logger, now Clock, ttl time.Duration) *Cache {
	return &Cache{
		now:      now,
		ttl:      ttl,
		logger:   logger,
		entries:  make(map[string]entry),
		inflight: make(map[string]*call),
	}
}

func (c *Cache) Get(ctx context.Context, key string, fetch func(context.Context) ([]feed.Item, error)) ([]feed.Item, error) {
	c.mu.Lock()
	cached, found := c.entries[key]
	if found && c.now().Before(cached.expires) {
		c.mu.Unlock()
		return cached.items, nil
	}

	if pending, running := c.inflight[key]; running {
		c.mu.Unlock()
		return c.wait(ctx, pending)
	}

	pending := &call{done: make(chan struct{})}
	c.inflight[key] = pending
	c.mu.Unlock()

	items, err := fetch(context.WithoutCancel(ctx))

	c.mu.Lock()
	switch {
	case err == nil:
		c.logger.Info("feed refreshed from upstream", "key", key, "items", len(items))
		c.entries[key] = entry{items: items, expires: c.now().Add(c.ttl)}
		pending.items = items
	case found:
		c.logger.Warn("feed upstream failed, serving stale", "key", key, "error", err)
		c.entries[key] = entry{items: cached.items, expires: c.now().Add(retryAfterErr)}
		pending.items = cached.items
	default:
		c.logger.Error("feed upstream failed, nothing cached", "key", key, "error", err)
		pending.err = err
	}
	delete(c.inflight, key)
	c.mu.Unlock()
	close(pending.done)

	return pending.items, pending.err
}

func (c *Cache) wait(ctx context.Context, pending *call) ([]feed.Item, error) {
	select {
	case <-pending.done:
		return pending.items, pending.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type CachedRepositoryFeed struct {
	cache *Cache
	inner feed.RepositoryFeed
}

func NewCachedRepositoryFeed(cache *Cache, inner feed.RepositoryFeed) *CachedRepositoryFeed {
	return &CachedRepositoryFeed{cache: cache, inner: inner}
}

func (f *CachedRepositoryFeed) Repositories(ctx context.Context) ([]feed.Item, error) {
	return f.cache.Get(ctx, "repositories", f.inner.Repositories)
}

type CachedPackageFeed struct {
	cache *Cache
	inner feed.PackageFeed
}

func NewCachedPackageFeed(cache *Cache, inner feed.PackageFeed) *CachedPackageFeed {
	return &CachedPackageFeed{cache: cache, inner: inner}
}

func (f *CachedPackageFeed) Packages(ctx context.Context) ([]feed.Item, error) {
	return f.cache.Get(ctx, "packages", f.inner.Packages)
}
