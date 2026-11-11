package query_handler

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/query"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/feed"
)

type fakeRepositoryFeed struct {
	items []feed.Item
	err   error
}

func (f fakeRepositoryFeed) Repositories(_ context.Context) ([]feed.Item, error) {
	return f.items, f.err
}

type fakePackageFeed struct {
	items []feed.Item
	err   error
}

func (f fakePackageFeed) Packages(_ context.Context) ([]feed.Item, error) {
	return f.items, f.err
}

func TestListRepositoriesHandler(t *testing.T) {
	t.Run("it returns the feed items", func(t *testing.T) {
		items := []feed.Item{{Name: "a", URL: "https://x"}}
		h := &ListRepositoriesHandler{Feed: fakeRepositoryFeed{items: items}}

		got, err := h.Handle(context.Background(), &query.ListRepositories{})

		require.NoError(t, err)
		assert.Equal(t, items, got)
	})

	t.Run("it returns an empty slice rather than nil when the feed is empty", func(t *testing.T) {
		h := &ListRepositoriesHandler{Feed: fakeRepositoryFeed{}}

		got, err := h.Handle(context.Background(), &query.ListRepositories{})

		require.NoError(t, err)
		assert.Equal(t, []feed.Item{}, got)
	})

	t.Run("it wraps the feed error", func(t *testing.T) {
		boom := errors.New("boom")
		h := &ListRepositoriesHandler{Feed: fakeRepositoryFeed{err: boom}}

		_, err := h.Handle(context.Background(), &query.ListRepositories{})

		assert.ErrorIs(t, err, boom)
	})

	t.Run("it rejects a query of another type", func(t *testing.T) {
		h := &ListRepositoriesHandler{Feed: fakeRepositoryFeed{}}

		_, err := h.Handle(context.Background(), &query.ListPackages{})

		assert.Error(t, err)
	})
}

func TestListPackagesHandler(t *testing.T) {
	t.Run("it returns the feed items", func(t *testing.T) {
		items := []feed.Item{{Name: "p"}}
		h := &ListPackagesHandler{Feed: fakePackageFeed{items: items}}

		got, err := h.Handle(context.Background(), &query.ListPackages{})

		require.NoError(t, err)
		assert.Equal(t, items, got)
	})

	t.Run("it wraps the feed error", func(t *testing.T) {
		boom := errors.New("boom")
		h := &ListPackagesHandler{Feed: fakePackageFeed{err: boom}}

		_, err := h.Handle(context.Background(), &query.ListPackages{})

		assert.ErrorIs(t, err, boom)
	})

	t.Run("it rejects a query of another type", func(t *testing.T) {
		h := &ListPackagesHandler{Feed: fakePackageFeed{}}

		_, err := h.Handle(context.Background(), &query.ListRepositories{})

		assert.Error(t, err)
	})
}
