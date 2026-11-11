package feed

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/feed"
)

const maxRepositories = 10

const GitHubReposURL = "https://api.github.com/users/pascalallen/repos?sort=updated&per_page=30&type=owner"

type GitHubClient struct {
	url    string
	token  string
	client *http.Client
}

func NewGitHubClient(url string, token string, client *http.Client) *GitHubClient {
	if client == nil {
		client = NewHTTPClient()
	}

	return &GitHubClient{url: url, token: token, client: client}
}

type gitHubRepository struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	HTMLURL     string  `json:"html_url"`
	Language    *string `json:"language"`
	Stars       int     `json:"stargazers_count"`
	Fork        bool    `json:"fork"`
	Archived    bool    `json:"archived"`
}

func (g *GitHubClient) Repositories(ctx context.Context) ([]feed.Item, error) {
	var repos []gitHubRepository
	if err := fetchJSON(ctx, g.client, g.url, g.token, &repos); err != nil {
		return nil, fmt.Errorf("fetch github repositories: %w", err)
	}

	items := make([]feed.Item, 0, len(repos))
	for _, r := range repos {
		if r.Fork || r.Archived {
			continue
		}
		if len(items) == maxRepositories {
			break
		}
		item := feed.Item{Name: r.Name, URL: r.HTMLURL}
		if r.Description != nil {
			item.Description = *r.Description
		}
		item.Meta = fmt.Sprintf("%d stars", r.Stars)
		if r.Language != nil && *r.Language != "" {
			item.Meta = *r.Language + ", " + item.Meta
		}
		items = append(items, item)
	}

	return items, nil
}
