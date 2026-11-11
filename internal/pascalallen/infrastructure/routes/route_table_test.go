package routes

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

var declaredRoutes = []string{
	"GET /",
	"GET /api/v1/camera/status",
	"GET /api/v1/camera/stream",
	"GET /api/v1/packages",
	"GET /api/v1/repos",
	"GET /plain",
	"GET /robots.txt",
	"GET /sitemap.xml",
	"GET /static/*filepath",
	"HEAD /",
	"HEAD /api/v1/camera/status",
	"HEAD /api/v1/packages",
	"HEAD /api/v1/repos",
	"HEAD /plain",
	"HEAD /robots.txt",
	"HEAD /sitemap.xml",
	"HEAD /static/*filepath",
	"POST /api/v1/visits",
}

func TestRouteTable(t *testing.T) {
	t.Run("the registered routes are exactly the declared list", func(t *testing.T) {
		var registered []string
		for _, route := range newTestRouter(t).engine.Routes() {
			registered = append(registered, route.Method+" "+route.Path)
		}
		sort.Strings(registered)
		declared := append([]string{}, declaredRoutes...)
		sort.Strings(declared)

		assert.Equal(t, declared, registered)
	})
}
