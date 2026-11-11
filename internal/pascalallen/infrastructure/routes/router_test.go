package routes

import (
	"encoding/base64"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/render"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/camera"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/content"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/messaging"
)

func assembledRouter(t *testing.T, stream camera.Stream) Router {
	t.Helper()
	gin.SetMode(gin.TestMode)
	if _, err := os.Stat("web/template"); err != nil {
		t.Chdir("../../../..")
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	return Assemble(Deps{
		Page:       render.NewPage(content.FS(), render.Config{BaseURL: os.Getenv("APP_BASE_URL"), Env: os.Getenv("APP_ENV")}),
		Camera:     stream,
		QueryBus:   messaging.NewSynchronousQueryBus(logger),
		CommandBus: messaging.NewSynchronousCommandBus(logger),
	})
}

func newTestRouter(t *testing.T) Router {
	t.Helper()

	return assembledRouter(t, stubStream{})
}

func get(r Router, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	r.engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))

	return rec
}

func TestRouter(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "ghp_secret_value_for_test")
	t.Setenv("SLACK_DM_URL", "https://hooks.example.test/secret-slack-path")
	t.Setenv("APP_BASE_URL", "http://localhost:9938")
	t.Setenv("APP_ENV", "test")
	r := newTestRouter(t)

	t.Run("the root path renders the page with the name", func(t *testing.T) {
		rec := get(r, "/")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "text/html")
		assert.Contains(t, rec.Body.String(), "Pascal Allen")
	})

	t.Run("the page carries neither server-side secret names nor values", func(t *testing.T) {
		body := get(r, "/").Body.String()

		assert.NotContains(t, body, "GITHUB_TOKEN")
		assert.NotContains(t, body, "SLACK_DM_URL")
		assert.NotContains(t, body, "ghp_secret_value_for_test")
		assert.NotContains(t, body, "secret-slack-path")
	})

	t.Run("the encoded page config holds only the base url, environment and asset hashes", func(t *testing.T) {
		body := get(r, "/").Body.String()
		m := regexp.MustCompile(`id="script_config" type="hidden" value="([^"]+)"`).FindStringSubmatch(body)
		require.Len(t, m, 2)
		raw, err := base64.StdEncoding.DecodeString(m[1])
		require.NoError(t, err)

		assert.JSONEq(t, `{"APP_BASE_URL":"http://localhost:9938","APP_ENV":"test","ASSET_HASHES":{}}`, string(raw))
	})

	t.Run("an unknown path renders the page marked noindex", func(t *testing.T) {
		rec := get(r, "/no/such/path")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `<meta name="robots" content="noindex">`)
		assert.Contains(t, rec.Body.String(), `data-prerendered=""`)
		assert.NotContains(t, get(r, "/").Body.String(), "noindex")
	})

	t.Run("a static asset is served", func(t *testing.T) {
		rec := get(r, "/static/favicon.ico")

		require.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Header().Get("Content-Type"), "text/html")
		assert.NotContains(t, rec.Body.String(), "Pascal Allen")
	})
}
