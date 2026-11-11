package feed

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/feed"
)

type testClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *testClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.now
}

func (c *testClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func newTestCache() (*Cache, *testClock) {
	clock := &testClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}

	return NewCacheWithClock(slog.New(slog.NewTextHandler(io.Discard, nil)), clock.Now, CacheTTL), clock
}

type countingFetch struct {
	calls atomic.Int32
	items []feed.Item
	err   error
}

func (f *countingFetch) fetch(_ context.Context) ([]feed.Item, error) {
	f.calls.Add(1)

	return f.items, f.err
}

func TestCache(t *testing.T) {
	items := []feed.Item{{Name: "a"}}

	t.Run("it serves a second call from the cache within the ttl", func(t *testing.T) {
		cache, clock := newTestCache()
		f := &countingFetch{items: items}

		_, err := cache.Get(context.Background(), "k", f.fetch)
		require.NoError(t, err)
		clock.Advance(CacheTTL - time.Second)
		got, err := cache.Get(context.Background(), "k", f.fetch)

		require.NoError(t, err)
		assert.Equal(t, items, got)
		assert.Equal(t, int32(1), f.calls.Load())
	})

	t.Run("it refetches after the ttl expires", func(t *testing.T) {
		cache, clock := newTestCache()
		f := &countingFetch{items: items}

		_, _ = cache.Get(context.Background(), "k", f.fetch)
		clock.Advance(CacheTTL)
		_, err := cache.Get(context.Background(), "k", f.fetch)

		require.NoError(t, err)
		assert.Equal(t, int32(2), f.calls.Load())
	})

	t.Run("it keeps keys independent", func(t *testing.T) {
		cache, _ := newTestCache()
		f := &countingFetch{items: items}

		_, _ = cache.Get(context.Background(), "a", f.fetch)
		_, _ = cache.Get(context.Background(), "b", f.fetch)

		assert.Equal(t, int32(2), f.calls.Load())
	})

	t.Run("it serves the stale value when the upstream fails after expiry", func(t *testing.T) {
		cache, clock := newTestCache()
		good := &countingFetch{items: items}
		bad := &countingFetch{err: errors.New("down")}

		_, _ = cache.Get(context.Background(), "k", good.fetch)
		clock.Advance(CacheTTL + time.Second)
		got, err := cache.Get(context.Background(), "k", bad.fetch)

		require.NoError(t, err)
		assert.Equal(t, items, got)
	})

	t.Run("it backs off the upstream after serving stale", func(t *testing.T) {
		cache, clock := newTestCache()
		good := &countingFetch{items: items}
		bad := &countingFetch{err: errors.New("down")}

		_, _ = cache.Get(context.Background(), "k", good.fetch)
		clock.Advance(CacheTTL + time.Second)
		_, _ = cache.Get(context.Background(), "k", bad.fetch)
		_, _ = cache.Get(context.Background(), "k", bad.fetch)

		assert.Equal(t, int32(1), bad.calls.Load())
		clock.Advance(retryAfterErr)
		_, _ = cache.Get(context.Background(), "k", bad.fetch)
		assert.Equal(t, int32(2), bad.calls.Load())
	})

	t.Run("it returns the error when nothing is cached", func(t *testing.T) {
		cache, _ := newTestCache()
		boom := errors.New("down")
		bad := &countingFetch{err: boom}

		_, err := cache.Get(context.Background(), "k", bad.fetch)

		assert.ErrorIs(t, err, boom)
	})

	t.Run("it does not cache an error", func(t *testing.T) {
		cache, _ := newTestCache()
		bad := &countingFetch{err: errors.New("down")}

		_, _ = cache.Get(context.Background(), "k", bad.fetch)
		_, _ = cache.Get(context.Background(), "k", bad.fetch)

		assert.Equal(t, int32(2), bad.calls.Load())
	})

	t.Run("it makes one upstream call for a burst of 50 goroutines", func(t *testing.T) {
		cache, _ := newTestCache()
		var calls atomic.Int32
		release := make(chan struct{})
		fetch := func(context.Context) ([]feed.Item, error) {
			calls.Add(1)
			<-release

			return items, nil
		}

		var wg sync.WaitGroup
		results := make([][]feed.Item, 50)
		for i := range 50 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				results[i], _ = cache.Get(context.Background(), "k", fetch)
			}()
		}
		require.Eventually(t, func() bool { return calls.Load() == 1 }, time.Second, time.Millisecond)
		time.Sleep(20 * time.Millisecond)
		close(release)
		wg.Wait()

		assert.Equal(t, int32(1), calls.Load())
		for _, r := range results {
			assert.Equal(t, items, r)
		}
	})

	t.Run("a waiting caller returns when its context is cancelled", func(t *testing.T) {
		cache, _ := newTestCache()
		release := make(chan struct{})
		started := make(chan struct{})
		go func() {
			_, _ = cache.Get(context.Background(), "k", func(context.Context) ([]feed.Item, error) {
				close(started)
				<-release

				return items, nil
			})
		}()
		<-started
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := cache.Get(ctx, "k", (&countingFetch{}).fetch)
		close(release)

		assert.ErrorIs(t, err, context.Canceled)
	})
}

type stubRepositories struct{ calls atomic.Int32 }

func (s *stubRepositories) Repositories(context.Context) ([]feed.Item, error) {
	s.calls.Add(1)

	return []feed.Item{{Name: "r"}}, nil
}

type stubPackages struct{ calls atomic.Int32 }

func (s *stubPackages) Packages(context.Context) ([]feed.Item, error) {
	s.calls.Add(1)

	return []feed.Item{{Name: "p"}}, nil
}

func TestCachedFeeds(t *testing.T) {
	t.Run("the repository and package decorators share a cache without colliding", func(t *testing.T) {
		cache, _ := newTestCache()
		repos, pkgs := &stubRepositories{}, &stubPackages{}
		cr, cp := NewCachedRepositoryFeed(cache, repos), NewCachedPackageFeed(cache, pkgs)

		r1, _ := cr.Repositories(context.Background())
		p1, _ := cp.Packages(context.Background())
		_, _ = cr.Repositories(context.Background())
		_, _ = cp.Packages(context.Background())

		assert.Equal(t, "r", r1[0].Name)
		assert.Equal(t, "p", p1[0].Name)
		assert.Equal(t, int32(1), repos.calls.Load())
		assert.Equal(t, int32(1), pkgs.calls.Load())
	})
}
