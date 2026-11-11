package feed

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const githubBody = `[
  {"name":"pgqueue","description":"queue","html_url":"https://github.com/pascalallen/pgqueue","language":"Go","stargazers_count":7},
  {"name":"nolang","description":null,"html_url":"https://github.com/pascalallen/nolang","language":null,"stargazers_count":0}
]`

const npmBody = `{"objects":[{"package":{"name":"@pascalallen/react-form-components","version":"2.1.0","description":"forms","links":{"npm":"https://www.npmjs.com/package/@pascalallen/react-form-components"}}}]}`

func serve(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	return srv
}

func TestGitHubClient(t *testing.T) {
	t.Run("it maps repositories to items", func(t *testing.T) {
		srv := serve(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(githubBody)) })

		items, err := NewGitHubClient(srv.URL, "", nil).Repositories(context.Background())

		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, "pgqueue", items[0].Name)
		assert.Equal(t, "queue", items[0].Description)
		assert.Equal(t, "https://github.com/pascalallen/pgqueue", items[0].URL)
		assert.Equal(t, "Go, 7 stars", items[0].Meta)
		assert.Equal(t, "", items[1].Description)
		assert.Equal(t, "0 stars", items[1].Meta)
	})

	t.Run("it drops forks and archived repositories", func(t *testing.T) {
		body := `[
  {"name":"keep","html_url":"u1","stargazers_count":1},
  {"name":"forked","html_url":"u2","fork":true},
  {"name":"old","html_url":"u3","archived":true},
  {"name":"keep2","html_url":"u4"}
]`
		srv := serve(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(body)) })

		items, err := NewGitHubClient(srv.URL, "", nil).Repositories(context.Background())

		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, "keep", items[0].Name)
		assert.Equal(t, "keep2", items[1].Name)
	})

	t.Run("it returns at most ten repositories after filtering", func(t *testing.T) {
		var parts []string
		for i := 0; i < 5; i++ {
			parts = append(parts, `{"name":"fork","fork":true}`)
		}
		for i := 0; i < 15; i++ {
			parts = append(parts, `{"name":"r"}`)
		}
		body := "[" + strings.Join(parts, ",") + "]"
		srv := serve(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(body)) })

		items, err := NewGitHubClient(srv.URL, "", nil).Repositories(context.Background())

		require.NoError(t, err)
		assert.Len(t, items, 10)
	})

	t.Run("the production url asks for 30 repositories sorted by updated", func(t *testing.T) {
		assert.Contains(t, GitHubReposURL, "per_page=30")
		assert.Contains(t, GitHubReposURL, "sort=updated")
		assert.Contains(t, GitHubReposURL, "type=owner")
	})

	t.Run("it sends the generic user agent and no authorization without a token", func(t *testing.T) {
		var got http.Header
		srv := serve(t, func(w http.ResponseWriter, r *http.Request) {
			got = r.Header.Clone()
			_, _ = w.Write([]byte("[]"))
		})

		_, err := NewGitHubClient(srv.URL, "", nil).Repositories(context.Background())

		require.NoError(t, err)
		assert.Equal(t, userAgent, got.Get("User-Agent"))
		assert.Empty(t, got.Get("Authorization"))
	})

	t.Run("it sends the token as a bearer when set", func(t *testing.T) {
		var auth string
		srv := serve(t, func(w http.ResponseWriter, r *http.Request) {
			auth = r.Header.Get("Authorization")
			_, _ = w.Write([]byte("[]"))
		})

		_, err := NewGitHubClient(srv.URL, "tok", nil).Repositories(context.Background())

		require.NoError(t, err)
		assert.Equal(t, "Bearer tok", auth)
	})

	t.Run("it fails on an upstream 500", func(t *testing.T) {
		srv := serve(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`[]`))
		})

		_, err := NewGitHubClient(srv.URL, "", nil).Repositories(context.Background())

		assert.Error(t, err)
	})

	t.Run("it fails when the upstream is slower than the client timeout", func(t *testing.T) {
		release := make(chan struct{})
		srv := serve(t, func(w http.ResponseWriter, _ *http.Request) { <-release })
		t.Cleanup(func() { close(release) })

		_, err := NewGitHubClient(srv.URL, "", &http.Client{Timeout: 50 * time.Millisecond}).Repositories(context.Background())

		assert.Error(t, err)
	})

	t.Run("it fails on a body over the size cap", func(t *testing.T) {
		srv := serve(t, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("[" + strings.Repeat(" ", maxResponseSize+1) + "]"))
		})

		_, err := NewGitHubClient(srv.URL, "", nil).Repositories(context.Background())

		assert.ErrorIs(t, err, ErrResponseTooLarge)
	})

	t.Run("it fails on malformed json", func(t *testing.T) {
		srv := serve(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("{not json")) })

		_, err := NewGitHubClient(srv.URL, "", nil).Repositories(context.Background())

		assert.Error(t, err)
	})
}

func TestNpmClient(t *testing.T) {
	t.Run("it maps packages to items", func(t *testing.T) {
		srv := serve(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(npmBody)) })

		items, err := NewNpmClient(srv.URL, nil).Packages(context.Background())

		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, "@pascalallen/react-form-components", items[0].Name)
		assert.Equal(t, "forms", items[0].Description)
		assert.Equal(t, "https://www.npmjs.com/package/@pascalallen/react-form-components", items[0].URL)
		assert.Equal(t, "v2.1.0", items[0].Meta)
	})

	t.Run("it never sends an authorization header", func(t *testing.T) {
		var auth string
		srv := serve(t, func(w http.ResponseWriter, r *http.Request) {
			auth = r.Header.Get("Authorization")
			_, _ = w.Write([]byte(`{"objects":[]}`))
		})

		_, err := NewNpmClient(srv.URL, nil).Packages(context.Background())

		require.NoError(t, err)
		assert.Empty(t, auth)
	})

	t.Run("it fails on an upstream 500", func(t *testing.T) {
		srv := serve(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"objects":[]}`))
		})

		_, err := NewNpmClient(srv.URL, nil).Packages(context.Background())

		assert.Error(t, err)
	})

	t.Run("it fails when the upstream is slower than the client timeout", func(t *testing.T) {
		release := make(chan struct{})
		srv := serve(t, func(w http.ResponseWriter, _ *http.Request) { <-release })
		t.Cleanup(func() { close(release) })

		_, err := NewNpmClient(srv.URL, &http.Client{Timeout: 50 * time.Millisecond}).Packages(context.Background())

		assert.Error(t, err)
	})

	t.Run("it fails on a body over the size cap", func(t *testing.T) {
		srv := serve(t, func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"objects":[]}` + strings.Repeat(" ", maxResponseSize)))
		})

		_, err := NewNpmClient(srv.URL, nil).Packages(context.Background())

		assert.ErrorIs(t, err, ErrResponseTooLarge)
	})

	t.Run("it fails on malformed json", func(t *testing.T) {
		srv := serve(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("<html>")) })

		_, err := NewNpmClient(srv.URL, nil).Packages(context.Background())

		assert.Error(t, err)
	})
}

func TestProductionTimeouts(t *testing.T) {
	t.Run("the default timeout is non-zero and used by the default client and both constructors", func(t *testing.T) {
		assert.Greater(t, DefaultTimeout, time.Duration(0))
		assert.Equal(t, DefaultTimeout, NewHTTPClient().Timeout)
		assert.Equal(t, DefaultTimeout, NewGitHubClient("http://x", "", nil).client.Timeout)
		assert.Equal(t, DefaultTimeout, NewNpmClient("http://x", nil).client.Timeout)
	})

	t.Run("clients built from a short timeout give up on a slow server", func(t *testing.T) {
		release := make(chan struct{})
		srv := serve(t, func(http.ResponseWriter, *http.Request) { <-release })
		t.Cleanup(func() { close(release) })

		start := time.Now()
		_, errGitHub := NewGitHubClient(srv.URL, "", NewHTTPClientWithTimeout(50*time.Millisecond)).Repositories(context.Background())
		_, errNpm := NewNpmClient(srv.URL, NewHTTPClientWithTimeout(50*time.Millisecond)).Packages(context.Background())

		assert.Error(t, errGitHub)
		assert.Error(t, errNpm)
		assert.Less(t, time.Since(start), 2*time.Second)
	})
}
