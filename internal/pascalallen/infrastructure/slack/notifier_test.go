package slack

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/visit"
)

func TestNotifier(t *testing.T) {
	v := visit.Visit{Language: "en-US", UserAgent: "Mozilla"}

	t.Run("it posts one short line in production when the url is set", func(t *testing.T) {
		var body map[string]string
		var method, contentType string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			method, contentType = r.Method, r.Header.Get("Content-Type")
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &body)
		}))
		t.Cleanup(srv.Close)

		err := NewNotifier(srv.URL, "production", nil).Notify(context.Background(), v)

		require.NoError(t, err)
		assert.Equal(t, http.MethodPost, method)
		assert.Equal(t, "application/json", contentType)
		assert.Contains(t, body["text"], "en-US")
		assert.Contains(t, body["text"], "Mozilla")
	})

	t.Run("it does nothing outside production", func(t *testing.T) {
		var calls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls.Add(1) }))
		t.Cleanup(srv.Close)

		err := NewNotifier(srv.URL, "local", nil).Notify(context.Background(), v)

		require.NoError(t, err)
		assert.Equal(t, int32(0), calls.Load())
	})

	t.Run("it does nothing when the url is empty", func(t *testing.T) {
		assert.NoError(t, NewNotifier("", "production", nil).Notify(context.Background(), v))
	})

	t.Run("it reports a non-2xx status without the url", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusInternalServerError) }))
		t.Cleanup(srv.Close)

		err := NewNotifier(srv.URL+"/secret-path", "production", nil).Notify(context.Background(), v)

		require.Error(t, err)
		assert.NotContains(t, err.Error(), "secret-path")
	})

	t.Run("it does not leak the webhook url on a transport error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		target := srv.URL + "/secret-path"
		srv.Close()

		err := NewNotifier(target, "production", &http.Client{Timeout: time.Second}).Notify(context.Background(), v)

		require.Error(t, err)
		assert.NotContains(t, err.Error(), "secret-path")
	})
}
