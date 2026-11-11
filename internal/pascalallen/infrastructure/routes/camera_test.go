package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/camera"
)

type stubStream struct {
	frames    chan []byte
	available bool
	released  chan struct{}
}

func (s stubStream) Subscribe(ctx context.Context) (<-chan []byte, func(), error) {
	if s.frames == nil {
		return nil, nil, camera.ErrUnavailable
	}

	return s.frames, func() { close(s.released) }, nil
}

func (s stubStream) Available() bool {
	return s.available
}

func cameraRouter(stream camera.Stream) Router {
	gin.SetMode(gin.TestMode)
	r := NewRouter()
	r.Camera(stream)

	return r
}

func TestCameraRoutes(t *testing.T) {
	t.Run("the stream is a 503 JSend error when the camera is unavailable", func(t *testing.T) {
		rec := get(cameraRouter(stubStream{}), "/api/v1/camera/stream")

		require.Equal(t, http.StatusServiceUnavailable, rec.Code)
		var body map[string]any
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		assert.Equal(t, "error", body["status"])
		assert.Equal(t, "camera unavailable", body["message"])
	})

	t.Run("the status reports unavailable", func(t *testing.T) {
		rec := get(cameraRouter(stubStream{}), "/api/v1/camera/status")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"status":"success","data":{"available":false}}`, rec.Body.String())
	})

	t.Run("the status reports available", func(t *testing.T) {
		rec := get(cameraRouter(stubStream{available: true}), "/api/v1/camera/status")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"status":"success","data":{"available":true}}`, rec.Body.String())
	})

	t.Run("the stream ends when the hub closes the subscription", func(t *testing.T) {
		frames := make(chan []byte)
		released := make(chan struct{})
		r := cameraRouter(stubStream{frames: frames, available: true, released: released})
		rec := httptest.NewRecorder()
		done := make(chan struct{})

		go func() {
			r.engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/camera/stream", nil))
			close(done)
		}()
		close(frames)

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("handler did not return after the source ended")
		}
	})

	t.Run("the stream writes frames as multipart and releases on client disconnect", func(t *testing.T) {
		frames := make(chan []byte, 1)
		released := make(chan struct{})
		r := cameraRouter(stubStream{frames: frames, available: true, released: released})
		ctx, cancel := context.WithCancel(t.Context())
		req := httptest.NewRequest(http.MethodGet, "/api/v1/camera/stream", nil).WithContext(ctx)
		rec := httptest.NewRecorder()
		done := make(chan struct{})

		go func() {
			r.engine.ServeHTTP(rec, req)
			close(done)
		}()
		frames <- []byte("--frame\r\nchunk")
		assert.Eventually(t, func() bool { return len(frames) == 0 }, 2*time.Second, 5*time.Millisecond)
		cancel()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("handler did not return after disconnect")
		}
		select {
		case <-released:
		default:
			t.Fatal("subscription was not released")
		}
		assert.Equal(t, "multipart/x-mixed-replace; boundary=frame", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Body.String(), "chunk")
	})
}
