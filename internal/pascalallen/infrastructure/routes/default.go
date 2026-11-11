package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/render"
)

func (r Router) Default(page *render.Page) {
	r.engine.GET("/plain", func(c *gin.Context) {
		c.HTML(http.StatusOK, "plain.tmpl", page.PlainView())
	})
	r.engine.GET("/robots.txt", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(page.Robots()))
	})
	r.engine.GET("/sitemap.xml", func(c *gin.Context) {
		body, err := page.Sitemap()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "application/xml; charset=utf-8", body)
	})
	r.engine.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", page.Index(c.Request.URL.Path))
	})
}
