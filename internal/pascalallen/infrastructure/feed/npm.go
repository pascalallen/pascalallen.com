package feed

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/feed"
)

const NpmSearchURL = "https://registry.npmjs.org/-/v1/search?text=maintainer:pascalallen&size=20"

type NpmClient struct {
	url    string
	client *http.Client
}

func NewNpmClient(url string, client *http.Client) *NpmClient {
	if client == nil {
		client = NewHTTPClient()
	}

	return &NpmClient{url: url, client: client}
}

type npmSearchResponse struct {
	Objects []struct {
		Package struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Description string `json:"description"`
			Links       struct {
				Npm string `json:"npm"`
			} `json:"links"`
		} `json:"package"`
	} `json:"objects"`
}

func (n *NpmClient) Packages(ctx context.Context) ([]feed.Item, error) {
	var res npmSearchResponse
	if err := fetchJSON(ctx, n.client, n.url, "", &res); err != nil {
		return nil, fmt.Errorf("fetch npm packages: %w", err)
	}

	items := make([]feed.Item, 0, len(res.Objects))
	for _, o := range res.Objects {
		items = append(items, feed.Item{
			Name:        o.Package.Name,
			Description: o.Package.Description,
			URL:         o.Package.Links.Npm,
			Meta:        "v" + o.Package.Version,
		})
	}

	return items, nil
}
