package routes

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func staticEngine(t *testing.T, files map[string]string) *gin.Engine {
	t.Helper()
	engine, _ := staticEngineAt(t, files)

	return engine
}

func staticEngineAt(t *testing.T, files map[string]string) (*gin.Engine, string) {
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

	return engine, root
}

func hashOf(body string) string {
	sum := sha256.Sum256([]byte(body))

	return hex.EncodeToString(sum[:])[:12]
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

	t.Run("a request carrying the content hash is cached as immutable", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm?v="+hashOf("RAW"), map[string]string{"Accept-Encoding": "br"})

		assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))
		assert.Equal(t, "br", rec.Header().Get("Content-Encoding"))
		assert.Equal(t, "Accept-Encoding", rec.Header().Get("Vary"))
	})

	t.Run("the hash is checked against the file being served", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/app.js?v="+hashOf("RAW"), nil)

		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	})

	t.Run("a stale hash must revalidate", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm?v="+hashOf("OLD"), nil)

		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	})

	t.Run("a junk version must revalidate", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm?v=x", nil)

		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	})

	t.Run("a request without a version must revalidate", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm", nil)

		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	})

	t.Run("an empty version must revalidate", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm?v=", nil)

		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	})

	t.Run("a file without a known hash is never immutable", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/plain.css?v=abc", nil)

		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	})

	t.Run("the ETag is derived from the content hash", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm", nil)

		assert.Equal(t, `"`+hashOf("RAW")+`"`, rec.Header().Get("ETag"))
	})

	t.Run("each precompressed variant has its own ETag", func(t *testing.T) {
		br := request(engine, http.MethodGet, "/static/assets/shell.wasm", map[string]string{"Accept-Encoding": "br"})
		gz := request(engine, http.MethodGet, "/static/assets/shell.wasm", map[string]string{"Accept-Encoding": "gzip"})

		assert.Equal(t, `"`+hashOf("RAW")+`-br"`, br.Header().Get("ETag"))
		assert.Equal(t, `"`+hashOf("RAW")+`-gzip"`, gz.Header().Get("ETag"))
		assert.Equal(t, "no-cache", br.Header().Get("Cache-Control"))
		assert.Equal(t, "Accept-Encoding", br.Header().Get("Vary"))
	})

	t.Run("a matching If-None-Match gets 304", func(t *testing.T) {
		etag := `"` + hashOf("RAW") + `-br"`

		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm", map[string]string{"Accept-Encoding": "br", "If-None-Match": etag})

		assert.Equal(t, http.StatusNotModified, rec.Code)
		assert.Empty(t, rec.Body.String())
	})

	t.Run("a stale If-None-Match gets the file", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/shell.wasm", map[string]string{"If-None-Match": `"old"`})

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "RAW", rec.Body.String())
	})

	t.Run("a HEAD request carries the headers and no body", func(t *testing.T) {
		rec := request(engine, http.MethodHead, "/static/assets/shell.wasm", map[string]string{"Accept-Encoding": "br"})

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

func TestAssetIndex(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "assets"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "assets", "app.js"), []byte("abc"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "assets", "app.js.br"), []byte("zzz"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(root, "favicon.ico"), []byte("i"), 0o644))
	index := newAssetIndex(root)

	t.Run("hashes a file by its path under the root", func(t *testing.T) {
		assert.Equal(t, "ba7816bf8f01", index.hash("assets/app.js"))
		assert.Equal(t, hashOf("i"), index.hash("favicon.ico"))
	})

	t.Run("gives no hash to precompressed siblings, directories and missing files", func(t *testing.T) {
		assert.Empty(t, index.hash("assets/app.js.br"))
		assert.Empty(t, index.hash("assets"))
		assert.Empty(t, index.hash("assets/nope.js"))
		assert.Empty(t, index.hash(""))
	})

	t.Run("re-hashes when the file changes", func(t *testing.T) {
		file := filepath.Join(root, "assets", "live.js")
		require.NoError(t, os.WriteFile(file, []byte("one"), 0o644))
		first := index.hash("assets/live.js")
		require.NoError(t, os.WriteFile(file, []byte("two"), 0o644))
		require.NoError(t, os.Chtimes(file, time.Now().Add(time.Hour), time.Now().Add(time.Hour)))

		assert.Equal(t, hashOf("one"), first)
		assert.Equal(t, hashOf("two"), index.hash("assets/live.js"))
	})

	t.Run("forgets a removed file", func(t *testing.T) {
		file := filepath.Join(root, "assets", "gone.js")
		require.NoError(t, os.WriteFile(file, []byte("x"), 0o644))
		require.NotEmpty(t, index.hash("assets/gone.js"))
		require.NoError(t, os.Remove(file))

		assert.Empty(t, index.hash("assets/gone.js"))
	})
}

func TestAssetURL(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "assets"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "assets", "app.js"), []byte("abc"), 0o644))
	asset := assetURL(newAssetIndex(root))

	t.Run("carries the content hash of a known asset", func(t *testing.T) {
		assert.Equal(t, "/static/assets/app.js?v=ba7816bf8f01", asset("app.js"))
	})

	t.Run("falls back to the bare path for an unknown asset", func(t *testing.T) {
		assert.Equal(t, "/static/assets/app.css", asset("app.css"))
	})

	t.Run("follows a rebuilt asset without a restart", func(t *testing.T) {
		file := filepath.Join(root, "assets", "app.js")
		require.NoError(t, os.WriteFile(file, []byte("abd"), 0o644))
		require.NoError(t, os.Chtimes(file, time.Now().Add(time.Hour), time.Now().Add(time.Hour)))

		assert.Equal(t, "/static/assets/app.js?v="+hashOf("abd"), asset("app.js"))
	})
}

func TestServeStaticFollowsRebuiltFiles(t *testing.T) {
	engine, root := staticEngineAt(t, map[string]string{"assets/app.js": "one"})
	file := filepath.Join(root, "assets", "app.js")
	first := request(engine, http.MethodGet, "/static/assets/app.js?v="+hashOf("one"), nil)
	require.Equal(t, "public, max-age=31536000, immutable", first.Header().Get("Cache-Control"))
	oldTag := first.Header().Get("ETag")
	require.NoError(t, os.WriteFile(file, []byte("two!"), 0o644))
	later := time.Now().Add(time.Hour)
	require.NoError(t, os.Chtimes(file, later, later))

	t.Run("an old ETag gets the new bytes and a new ETag", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/app.js", map[string]string{"If-None-Match": oldTag})

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "two!", rec.Body.String())
		assert.Equal(t, `"`+hashOf("two!")+`"`, rec.Header().Get("ETag"))
	})

	t.Run("the old version is no longer immutable", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/app.js?v="+hashOf("one"), nil)

		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
	})

	t.Run("the new version is immutable", func(t *testing.T) {
		rec := request(engine, http.MethodGet, "/static/assets/app.js?v="+hashOf("two!"), nil)

		assert.Equal(t, "public, max-age=31536000, immutable", rec.Header().Get("Cache-Control"))
	})

	t.Run("a file added after startup is served with its hash", func(t *testing.T) {
		require.NoError(t, os.WriteFile(filepath.Join(root, "assets", "late.js"), []byte("late"), 0o644))

		rec := request(engine, http.MethodGet, "/static/assets/late.js", nil)

		assert.Equal(t, `"`+hashOf("late")+`"`, rec.Header().Get("ETag"))
	})

	t.Run("a removed file is a 404", func(t *testing.T) {
		require.NoError(t, os.Remove(file))

		assert.Equal(t, http.StatusNotFound, request(engine, http.MethodGet, "/static/assets/app.js", nil).Code)
	})

	t.Run("concurrent requests during rewrites are race free", func(t *testing.T) {
		var wg sync.WaitGroup
		target := filepath.Join(root, "assets", "busy.js")
		require.NoError(t, os.WriteFile(target, []byte("0"), 0o644))
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 20; j++ {
					request(engine, http.MethodGet, "/static/assets/busy.js", nil)
				}
			}()
		}
		for j := 1; j <= 20; j++ {
			require.NoError(t, os.WriteFile(target, []byte(strings.Repeat("x", j)), 0o644))
		}
		wg.Wait()

		rec := request(engine, http.MethodGet, "/static/assets/busy.js", nil)
		assert.Equal(t, `"`+hashOf(strings.Repeat("x", 20))+`"`, rec.Header().Get("ETag"))
	})
}

func TestConfigRegistersTheAssetFunction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Chdir("../../../..")
	r := NewRouter()
	r.Config()

	assert.Contains(t, r.engine.FuncMap, "asset")
}
