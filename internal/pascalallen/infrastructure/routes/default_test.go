package routes

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var versionedAsset = regexp.MustCompile(`(?:src|href|data-wasm-src)="(/static/assets/[^"?]+)\?v=([0-9a-f]{12})"`)

func ensureAsset(t *testing.T, name string) {
	t.Helper()
	path := filepath.Join("web", "static", "assets", name)
	if _, err := os.Stat(path); err == nil {
		return
	}
	require.NoError(t, os.WriteFile(path, []byte("stub "+name), 0o644))
	t.Cleanup(func() { _ = os.Remove(path) })
}

func versioned(t *testing.T, body, attribute, name string) string {
	t.Helper()
	match := regexp.MustCompile(attribute + `="(/static/assets/` + regexp.QuoteMeta(name) + `\?v=[0-9a-f]{12})"`).FindStringSubmatch(body)
	require.NotNil(t, match, "%s=%q with a version in the page", attribute, name)

	return match[1]
}

func TestDefaultPageAssets(t *testing.T) {
	t.Chdir("../../../..")
	require.NoError(t, os.MkdirAll("web/static/assets", 0o755))
	for _, name := range []string{"shell.wasm", "wasm_exec.js", "theme.js", "app.js", "app.css"} {
		ensureAsset(t, name)
	}

	rec := get(newTestRouter(t), "/anything")
	require.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	head := body[:strings.Index(body, "</head>")]

	t.Run("the page references the host assets through the asset helper", func(t *testing.T) {
		assert.Contains(t, body, `data-wasm-src="/static/assets/shell.wasm?v=`)
		assert.True(t, versionedAsset.MatchString(body))
	})

	t.Run("the runtime URL comes from a server-rendered attribute and is never a script tag", func(t *testing.T) {
		versioned(t, body, "data-runtime-src", "wasm_exec.js")
		assert.NotRegexp(t, `<script[^>]+wasm_exec\.js`, body)
	})

	t.Run("the module is not preloaded because the host starts its fetch itself", func(t *testing.T) {
		versioned(t, body, "data-wasm-src", "shell.wasm")

		assert.NotContains(t, head, `rel="preload"`)
		assert.NotContains(t, head, `<link href="/static/assets/shell.wasm`)
	})

	t.Run("the theme script blocks in the head before the stylesheet", func(t *testing.T) {
		theme := versioned(t, head, "src", "theme.js")
		tag := regexp.MustCompile(`<script src="` + regexp.QuoteMeta(theme) + `"></script>`)

		require.Regexp(t, tag, head)
		assert.Less(t, strings.Index(head, theme), strings.Index(head, "app.css"))
	})

	t.Run("the plain view carries the theme script too", func(t *testing.T) {
		plain := get(newTestRouter(t), "/plain").Body.String()

		versioned(t, plain, "src", "theme.js")
	})
}
