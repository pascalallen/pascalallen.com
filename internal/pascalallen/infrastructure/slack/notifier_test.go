package slack

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

	t.Run("it escapes slack markup and puts user values in code spans", func(t *testing.T) {
		var body map[string]string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &body)
		}))
		t.Cleanup(srv.Close)
		evil := visit.Visit{Language: "en`<!here>", UserAgent: "<https://evil.example|Reset your Slack password> <!channel> & ‮"}

		require.NoError(t, NewNotifier(srv.URL, "production", nil).Notify(context.Background(), evil))

		text := body["text"]
		assert.NotContains(t, text, "<")
		assert.NotContains(t, text, ">")
		assert.NotContains(t, text, "‮")
		assert.Contains(t, text, "&lt;https://evil.example|Reset your Slack password&gt; &lt;!channel&gt; &amp;")
		assert.Contains(t, text, "ua=`&lt;https://evil.example")
		assert.Equal(t, 4, strings.Count(text, "`"))
	})

	t.Run("it strips control and line separator characters from user values", func(t *testing.T) {
		var body map[string]string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(raw, &body)
		}))
		t.Cleanup(srv.Close)

		require.NoError(t, NewNotifier(srv.URL, "production", nil).Notify(context.Background(), visit.Visit{Language: "a\nb\x00c", UserAgent: "x\u2028y\u2029z"}))

		assert.Equal(t, "pascalallen.com visit: lang=`abc` ua=`xyz`", body["text"])
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

	t.Run("the production client uses the non-zero default timeout", func(t *testing.T) {
		assert.Greater(t, DefaultTimeout, time.Duration(0))
		assert.Equal(t, DefaultTimeout, NewNotifier("https://hooks.example.test/x", "production", nil).client.Timeout)
	})

	t.Run("a notifier built with a short timeout gives up on a slow server", func(t *testing.T) {
		release := make(chan struct{})
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { <-release }))
		t.Cleanup(srv.Close)
		t.Cleanup(func() { close(release) })

		start := time.Now()
		err := NewNotifierWithTimeout(srv.URL, "production", 50*time.Millisecond).Notify(context.Background(), v)

		require.Error(t, err)
		assert.Less(t, time.Since(start), 2*time.Second)
	})
}
