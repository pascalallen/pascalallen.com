package routes

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pageRouter(t *testing.T, module string) Router {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.SetHTMLTemplate(template.Must(template.New("index.tmpl").Parse("version=[{{ .wasmVersion }}]")))
	r := Router{engine: engine}
	r.defaultFor(module)

	return r
}

func TestDefaultPageWasmVersion(t *testing.T) {
	t.Run("exposes the content hash of the module to the template", func(t *testing.T) {
		module := filepath.Join(t.TempDir(), "shell.wasm")
		require.NoError(t, os.WriteFile(module, []byte("abc"), 0o644))

		rec := get(pageRouter(t, module), "/anything")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "version=[ba7816bf8f01]", rec.Body.String())
	})

	t.Run("renders with an empty version when the module is absent", func(t *testing.T) {
		rec := get(pageRouter(t, filepath.Join(t.TempDir(), "missing.wasm")), "/anything")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "version=[]", rec.Body.String())
	})
}
