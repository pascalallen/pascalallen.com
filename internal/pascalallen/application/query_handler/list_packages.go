package query_handler

import (
	"context"
	"fmt"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/messaging"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/query"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/feed"
)

type ListPackagesHandler struct {
	Feed feed.PackageFeed
}

func (h *ListPackagesHandler) Handle(ctx context.Context, q messaging.Query) (any, error) {
	if _, ok := q.(*query.ListPackages); !ok {
		return nil, fmt.Errorf("invalid query type: %T", q)
	}

	items, err := h.Feed.Packages(ctx)
	if err != nil {
		return nil, fmt.Errorf("list packages: %w", err)
	}

	return nonNil(items), nil
}
