package query_handler

import (
	"context"
	"fmt"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/messaging"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/query"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/feed"
)

type ListRepositoriesHandler struct {
	Feed feed.RepositoryFeed
}

func (h *ListRepositoriesHandler) Handle(ctx context.Context, q messaging.Query) (any, error) {
	if _, ok := q.(*query.ListRepositories); !ok {
		return nil, fmt.Errorf("invalid query type: %T", q)
	}

	items, err := h.Feed.Repositories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}

	return nonNil(items), nil
}

func nonNil(items []feed.Item) []feed.Item {
	if items == nil {
		return []feed.Item{}
	}

	return items
}
