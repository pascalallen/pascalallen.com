package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var versionedAsset = regexp.MustCompile(`(?:src|href|data-wasm-src)="(/static/assets/[^"?]+)\?v=([0-9a-f]{12})"`)

func TestDefaultPageAssets(t *testing.T) {
	t.Run("the page references the host assets through the asset helper", func(t *testing.T) {
		t.Chdir("../../../..")
		require.NoError(t, os.MkdirAll("web/static/assets", 0o755))
		module := filepath.Join("web", "static", "assets", "shell.wasm")
		if _, err := os.Stat(module); err != nil {
			require.NoError(t, os.WriteFile(module, []byte("abc"), 0o644))
			t.Cleanup(func() { _ = os.Remove(module) })
		}

		rec := get(newTestRouter(t), "/anything")

		require.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, `data-wasm-src="/static/assets/shell.wasm?v=`)
		assert.NotContains(t, body, "wasm_exec.js")
		assert.True(t, versionedAsset.MatchString(body))
	})
}
