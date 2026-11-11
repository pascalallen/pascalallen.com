package feed

import "context"

type Item struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Meta        string `json:"meta"`
}

type RepositoryFeed interface {
	Repositories(ctx context.Context) ([]Item, error)
}

type PackageFeed interface {
	Packages(ctx context.Context) ([]Item, error)
}
