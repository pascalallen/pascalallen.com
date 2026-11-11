package container

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/feed"
)

func TestNewConfig(t *testing.T) {
	t.Run("the feed URLs default to the public upstreams", func(t *testing.T) {
		t.Setenv("GITHUB_REPOS_URL", "")
		t.Setenv("NPM_SEARCH_URL", "")

		config := NewConfig()

		assert.Equal(t, feed.GitHubReposURL, config.GitHubReposURL)
		assert.Equal(t, feed.NpmSearchURL, config.NpmSearchURL)
	})

	t.Run("the feed URLs can be pointed at another server", func(t *testing.T) {
		t.Setenv("GITHUB_REPOS_URL", "http://127.0.0.1:1/github")
		t.Setenv("NPM_SEARCH_URL", "http://127.0.0.1:1/npm")

		config := NewConfig()

		assert.Equal(t, "http://127.0.0.1:1/github", config.GitHubReposURL)
		assert.Equal(t, "http://127.0.0.1:1/npm", config.NpmSearchURL)
	})
}

func TestFeedsUseTheConfiguredURLs(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/github":
			_, _ = w.Write([]byte(`[{"name":"fixture-repo","html_url":"https://github.com/x/fixture-repo","stargazers_count":1}]`))
		case "/npm":
			_, _ = w.Write([]byte(`{"objects":[{"package":{"name":"fixture-pkg","version":"1.0.0","links":{"npm":"https://www.npmjs.com/package/fixture-pkg"}}}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer upstream.Close()
	config := Config{GitHubReposURL: upstream.URL + "/github", NpmSearchURL: upstream.URL + "/npm"}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("the repository feed calls the configured URL", func(t *testing.T) {
		items, err := NewRepositoryFeed(config, feed.NewCache(logger)).Repositories(t.Context())

		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, "fixture-repo", items[0].Name)
	})

	t.Run("the package feed calls the configured URL", func(t *testing.T) {
		items, err := NewPackageFeed(config, feed.NewCache(logger)).Packages(t.Context())

		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, "fixture-pkg", items[0].Name)
	})
}
