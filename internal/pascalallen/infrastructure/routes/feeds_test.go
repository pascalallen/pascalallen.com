package routes

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/command"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/command_handler"
	appmessaging "github.com/pascalallen/pascalallen.com/internal/pascalallen/application/messaging"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/query"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/query_handler"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/feed"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/visit"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/messaging"
)

type stubRepositories struct {
	items []feed.Item
	err   error
}

func (s stubRepositories) Repositories(context.Context) ([]feed.Item, error) { return s.items, s.err }

type stubPackages struct {
	items []feed.Item
	err   error
}

func (s stubPackages) Packages(context.Context) ([]feed.Item, error) { return s.items, s.err }

type recordingNotifier struct {
	mu     sync.Mutex
	visits []visit.Visit
}

func (n *recordingNotifier) Notify(_ context.Context, v visit.Visit) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.visits = append(n.visits, v)

	return nil
}

type wrongTypeHandler struct{}

func (wrongTypeHandler) Handle(context.Context, appmessaging.Query) (any, error) {
	return "not items", nil
}

func newFeedsRouter(repos stubRepositories, pkgs stubPackages, notifier *recordingNotifier) Router {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	queryBus := messaging.NewSynchronousQueryBus(logger)
	queryBus.RegisterHandler((&query.ListRepositories{}).QueryName(), &query_handler.ListRepositoriesHandler{Feed: repos})
	queryBus.RegisterHandler((&query.ListPackages{}).QueryName(), &query_handler.ListPackagesHandler{Feed: pkgs})
	commandBus := messaging.NewSynchronousCommandBus(logger)
	commandBus.RegisterHandler((&command.RecordVisit{}).CommandName(), &command_handler.RecordVisitHandler{Notifier: notifier, Now: time.Now})

	r := NewRouter()
	r.Feeds(queryBus, commandBus)

	return r
}

func do(r Router, method string, path string, body string, remoteAddr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if remoteAddr != "" {
		req.RemoteAddr = remoteAddr
	}
	rec := httptest.NewRecorder()
	r.engine.ServeHTTP(rec, req)

	return rec
}

func TestFeedRoutes(t *testing.T) {
	items := []feed.Item{{Name: "pgqueue", Description: "queue", URL: "https://github.com/pascalallen/pgqueue", Meta: "Go, 7 stars"}}

	t.Run("GET repos returns a jsend success with items and a cache header", func(t *testing.T) {
		r := newFeedsRouter(stubRepositories{items: items}, stubPackages{}, &recordingNotifier{})

		rec := do(r, http.MethodGet, "/api/v1/repos", "", "")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "public, max-age=300", rec.Header().Get("Cache-Control"))
		var body struct {
			Status string `json:"status"`
			Data   struct {
				Items []feed.Item `json:"items"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "success", body.Status)
		assert.Equal(t, items, body.Data.Items)
	})

	t.Run("GET packages returns a jsend success with items and a cache header", func(t *testing.T) {
		r := newFeedsRouter(stubRepositories{}, stubPackages{items: items}, &recordingNotifier{})

		rec := do(r, http.MethodGet, "/api/v1/packages", "", "")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "public, max-age=300", rec.Header().Get("Cache-Control"))
		assert.Contains(t, rec.Body.String(), `"items":[{"name":"pgqueue"`)
	})

	t.Run("an empty feed serialises as an empty array", func(t *testing.T) {
		r := newFeedsRouter(stubRepositories{}, stubPackages{}, &recordingNotifier{})

		rec := do(r, http.MethodGet, "/api/v1/repos", "", "")

		assert.Contains(t, rec.Body.String(), `"items":[]`)
	})

	t.Run("a feed failure is a 502 jsend error that does not echo the upstream error", func(t *testing.T) {
		r := newFeedsRouter(stubRepositories{err: errors.New("dial tcp secret-host:443 refused")}, stubPackages{err: errors.New("boom")}, &recordingNotifier{})

		for _, path := range []string{"/api/v1/repos", "/api/v1/packages"} {
			rec := do(r, http.MethodGet, path, "", "")

			require.Equal(t, http.StatusBadGateway, rec.Code)
			assert.Empty(t, rec.Header().Get("Cache-Control"))
			assert.JSONEq(t, `{"status":"error","message":"feed unavailable","code":502}`, rec.Body.String())
			assert.NotContains(t, rec.Body.String(), "secret-host")
		}
	})

	t.Run("a handler result of an unexpected type is a 502", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		logger := slog.New(slog.NewTextHandler(io.Discard, nil))
		queryBus := messaging.NewSynchronousQueryBus(logger)
		queryBus.RegisterHandler((&query.ListRepositories{}).QueryName(), wrongTypeHandler{})
		r := NewRouter()
		r.Feeds(queryBus, messaging.NewSynchronousCommandBus(logger))

		rec := do(r, http.MethodGet, "/api/v1/repos", "", "")

		assert.Equal(t, http.StatusBadGateway, rec.Code)
	})

	t.Run("POST visits returns 204 and notifies", func(t *testing.T) {
		n := &recordingNotifier{}
		r := newFeedsRouter(stubRepositories{}, stubPackages{}, n)

		rec := do(r, http.MethodPost, "/api/v1/visits", `{"language":"en-US","userAgent":"ua"}`, "")

		require.Equal(t, http.StatusNoContent, rec.Code)
		assert.Empty(t, rec.Body.String())
		require.Len(t, n.visits, 1)
		assert.Equal(t, visit.Visit{Language: "en-US", UserAgent: "ua"}, n.visits[0])
	})

	t.Run("POST visits accepts an empty object and an empty body", func(t *testing.T) {
		r := newFeedsRouter(stubRepositories{}, stubPackages{}, &recordingNotifier{})

		assert.Equal(t, http.StatusNoContent, do(r, http.MethodPost, "/api/v1/visits", `{}`, "1.1.1.1:1").Code)
		assert.Equal(t, http.StatusNoContent, do(r, http.MethodPost, "/api/v1/visits", ``, "2.2.2.2:1").Code)
	})

	t.Run("POST visits falls back to the request headers", func(t *testing.T) {
		n := &recordingNotifier{}
		r := newFeedsRouter(stubRepositories{}, stubPackages{}, n)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/visits", strings.NewReader(`{}`))
		req.Header.Set("Accept-Language", "fr")
		req.Header.Set("User-Agent", "curl/8")

		r.engine.ServeHTTP(httptest.NewRecorder(), req)

		require.Len(t, n.visits, 1)
		assert.Equal(t, visit.Visit{Language: "fr", UserAgent: "curl/8"}, n.visits[0])
	})

	t.Run("POST visits rejects a body over 1KB with 413", func(t *testing.T) {
		n := &recordingNotifier{}
		r := newFeedsRouter(stubRepositories{}, stubPackages{}, n)

		rec := do(r, http.MethodPost, "/api/v1/visits", `{"language":"`+strings.Repeat("a", 2000)+`"}`, "")

		require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
		assert.Empty(t, n.visits)
	})

	t.Run("POST visits rejects malformed json with 400", func(t *testing.T) {
		n := &recordingNotifier{}
		r := newFeedsRouter(stubRepositories{}, stubPackages{}, n)

		rec := do(r, http.MethodPost, "/api/v1/visits", `{nope`, "")

		require.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Empty(t, n.visits)
	})

	t.Run("POST visits notifies once per client address and still returns 204 when limited", func(t *testing.T) {
		n := &recordingNotifier{}
		r := newFeedsRouter(stubRepositories{}, stubPackages{}, n)

		first := do(r, http.MethodPost, "/api/v1/visits", `{}`, "9.9.9.9:1000")
		second := do(r, http.MethodPost, "/api/v1/visits", `{}`, "9.9.9.9:2000")
		other := do(r, http.MethodPost, "/api/v1/visits", `{}`, "8.8.8.8:1000")

		assert.Equal(t, http.StatusNoContent, first.Code)
		assert.Equal(t, http.StatusNoContent, second.Code)
		assert.Equal(t, http.StatusNoContent, other.Code)
		assert.Len(t, n.visits, 2)
	})

	t.Run("responses carry no secret headers", func(t *testing.T) {
		r := newFeedsRouter(stubRepositories{items: items}, stubPackages{}, &recordingNotifier{})

		rec := do(r, http.MethodGet, "/api/v1/repos", "", "")

		assert.Empty(t, rec.Header().Get("Authorization"))
		assert.Empty(t, rec.Header().Get("Set-Cookie"))
	})
}

func TestTrustedProxies(t *testing.T) {
	forwarded := func(r Router, remote string, clientIP string) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/visits", strings.NewReader(`{}`))
		req.RemoteAddr = remote
		req.Header.Set("X-Forwarded-For", clientIP)
		r.engine.ServeHTTP(httptest.NewRecorder(), req)
	}

	t.Run("by default the forwarded header is ignored so every visitor behind the proxy shares one address", func(t *testing.T) {
		n := &recordingNotifier{}
		r := newFeedsRouter(stubRepositories{}, stubPackages{}, n)
		require.NoError(t, r.trustProxies(nil))

		forwarded(r, "10.0.0.1:1", "5.5.5.5")
		forwarded(r, "10.0.0.1:1", "6.6.6.6")

		assert.Len(t, n.visits, 1)
	})

	t.Run("a configured trusted proxy lets the forwarded address identify the visitor", func(t *testing.T) {
		n := &recordingNotifier{}
		r := newFeedsRouter(stubRepositories{}, stubPackages{}, n)
		require.NoError(t, r.trustProxies([]string{"10.0.0.0/8"}))

		forwarded(r, "10.0.0.1:1", "5.5.5.5")
		forwarded(r, "10.0.0.1:1", "6.6.6.6")
		forwarded(r, "10.0.0.1:1", "6.6.6.6")

		assert.Len(t, n.visits, 2)
	})

	t.Run("a forwarded header from an untrusted peer is ignored", func(t *testing.T) {
		n := &recordingNotifier{}
		r := newFeedsRouter(stubRepositories{}, stubPackages{}, n)
		require.NoError(t, r.trustProxies([]string{"10.0.0.0/8"}))

		forwarded(r, "203.0.113.9:1", "5.5.5.5")
		forwarded(r, "203.0.113.9:1", "6.6.6.6")

		assert.Len(t, n.visits, 1)
	})

	t.Run("it parses comma separated cidrs and addresses and skips blanks", func(t *testing.T) {
		got, err := ParseTrustedProxies(" 10.0.0.0/8, 192.168.1.5 ,,")

		require.NoError(t, err)
		assert.Equal(t, []string{"10.0.0.0/8", "192.168.1.5"}, got)
	})

	t.Run("an empty value trusts nothing", func(t *testing.T) {
		got, err := ParseTrustedProxies("")

		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("it rejects a malformed entry", func(t *testing.T) {
		_, err := ParseTrustedProxies("10.0.0.0/8,not-a-cidr")

		assert.Error(t, err)
	})
}
