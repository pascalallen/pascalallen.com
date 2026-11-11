package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func staticEngine(t *testing.T, files map[string]string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	root := filepath.Join(t.TempDir(), "static")
	require.NoError(t, os.MkdirAll(root, 0o755))
	for name, body := range files {
		full := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(body), 0o644))
	}
	engine := gin.New()
	handler := serveStatic(root)
	engine.GET("/static/*filepath", handler)
	engine.HEAD("/static/*filepath", handler)

	return engine
}

func request(engine *gin.Engine, method, target string, header map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	return rec
}

func TestServeStatic(t *testing.T) {
	files := map[string]string{
		"assets/shell.wasm":    "RAW",
		"assets/shell.wasm.br": "BROTLI",
		"assets/shell.wasm.gz": "GZIP",
		"assets/only.wasm":     "ONLY",
		"assets/app.js":        "console.log(1)",
		"assets/app.js.gz":     "GZJS",
		"assets/plain.css":     "body{}",
		"../secret.txt":        "outside",
	}
	engine := staticEngine(t, files)

	t.Run("a wasm file is served as application/wasm", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/only.wasm", nil)

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/wasm", rec.Header().Get("Content-Type"))
		assert.Equal(t, "ONLY", rec.Body.String())
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
	})

	t.Run("brotli is served when accepted", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm", map[string]string{"Accept-Encoding": "br"})

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "BROTLI", rec.Body.String())
		assert.Equal(t, "br", rec.Header().Get("Content-Encoding"))
		assert.Equal(t, "application/wasm", rec.Header().Get("Content-Type"))
		assert.Equal(t, "Accept-Encoding", rec.Header().Get("Vary"))
	})

	t.Run("brotli wins over gzip when both are accepted", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm", map[string]string{"Accept-Encoding": "gzip, deflate, br"})

		assert.Equal(t, "BROTLI", rec.Body.String())
		assert.Equal(t, "br", rec.Header().Get("Content-Encoding"))
	})

	t.Run("gzip is served when brotli is not accepted", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm", map[string]string{"Accept-Encoding": "gzip, deflate"})

		assert.Equal(t, "GZIP", rec.Body.String())
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
		assert.Equal(t, "application/wasm", rec.Header().Get("Content-Type"))
	})

	t.Run("an encoding refused with q=0 is not served", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm", map[string]string{"Accept-Encoding": "br;q=0, gzip;q=0.5"})

		assert.Equal(t, "GZIP", rec.Body.String())
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
	})

	t.Run("a refused gzip and no brotli serves the raw file", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm", map[string]string{"Accept-Encoding": "gzip;q=0"})

		assert.Equal(t, "RAW", rec.Body.String())
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
		assert.Equal(t, "Accept-Encoding", rec.Header().Get("Vary"))
	})

	t.Run("no Accept-Encoding serves the raw file", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm", nil)

		assert.Equal(t, "RAW", rec.Body.String())
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
	})

	t.Run("only the sibling that exists is used", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/app.js", map[string]string{"Accept-Encoding": "br, gzip"})

		assert.Equal(t, "GZJS", rec.Body.String())
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))
		assert.Contains(t, rec.Header().Get("Content-Type"), "javascript")
	})

	t.Run("a file without siblings has no Vary header", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/plain.css", map[string]string{"Accept-Encoding": "br"})

		assert.Equal(t, "body{}", rec.Body.String())
		assert.Empty(t, rec.Header().Get("Vary"))
		assert.Empty(t, rec.Header().Get("Content-Encoding"))
		assert.Contains(t, rec.Header().Get("Content-Type"), "text/css")
	})

	t.Run("a versioned request is cached as immutable", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm?v=abc123", map[string]string{"Accept-Encoding": "br"})

		assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))
	})

	t.Run("an unversioned request must revalidate", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm", nil)

		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	})

	t.Run("an empty version must revalidate", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm?v=", nil)

		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	})

	t.Run("a HEAD request carries the headers and no body", func(t *testing.T) {
		rec := request(engine, http.MethodHead, "/static/assets/shell.wasm?v=x", map[string]string{"Accept-Encoding": "br"})

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "br", rec.Header().Get("Content-Encoding"))
		assert.Equal(t, "application/wasm", rec.Header().Get("Content-Type"))
		assert.Equal(t, "6", rec.Header().Get("Content-Length"))
		assert.Empty(t, rec.Body.String())
	})

	t.Run("a missing file is a 404", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/missing.wasm", map[string]string{"Accept-Encoding": "br"})

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("a precompressed sibling without its original is a 404 for the original", func(t *testing.T) {
		engine := staticEngine(t, map[string]string{"assets/orphan.wasm.br": "BR"})

		rec := request(engine, http.MethodGet, "/static/assets/orphan.wasm", map[string]string{"Accept-Encoding": "br"})

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("a directory is a 404", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/", nil)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("path traversal cannot leave the static root", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/../../secret.txt", nil)
		assert.NotEqual(t, http.StatusOK, rec.Code)

		rec = request(engine, http.MethodGet, "/static/..%2fsecret.txt", nil)
		assert.NotEqual(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "outside")
	})

	t.Run("a conditional request gets 304", func(t *testing.T) {
		first := request(engine, http.MethodGet, "/static/assets/plain.css", nil)
		since := first.Header().Get("Last-Modified")
		require.NotEmpty(t, since)

		rec := request(engine, http.MethodGet, "/static/assets/plain.css", map[string]string{"If-Modified-Since": since})

		assert.Equal(t, http.StatusNotModified, rec.Code)
	})
}

func TestAssetVersion(t *testing.T) {
	t.Run("is a short hex digest of the file content", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "shell.wasm")
		require.NoError(t, os.WriteFile(path, []byte("abc"), 0o644))

		assert.Equal(t, "ba7816bf8f01", assetVersion(path))
	})

	t.Run("changes when the content changes", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "shell.wasm")
		require.NoError(t, os.WriteFile(path, []byte("abc"), 0o644))
		before := assetVersion(path)
		require.NoError(t, os.WriteFile(path, []byte("abd"), 0o644))

		assert.NotEqual(t, before, assetVersion(path))
	})

	t.Run("is empty when the file is absent", func(t *testing.T) {
		assert.Empty(t, assetVersion(filepath.Join(t.TempDir(), "missing.wasm")))
	})
}
