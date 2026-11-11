package action

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/responder"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/render"
)

const (
	htmlContentType = "text/html; charset=utf-8"
	allowedMethods  = "GET, HEAD"
)

func isHead(c *gin.Context) bool {
	return c.Request.Method == http.MethodHead
}

func headOnly(c *gin.Context, status int, contentType string) {
	c.Header("Content-Type", contentType)
	c.Status(status)
	c.Writer.WriteHeaderNow()
}

func HandleDefault(page *render.Page) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isHead(c) {
			headOnly(c, http.StatusOK, htmlContentType)
			return
		}
		c.HTML(http.StatusOK, "index.tmpl", page.Index(c.Request.URL.Path))
	}
}

func HandlePlain(page *render.Page) gin.HandlerFunc {
	return func(c *gin.Context) {
		if isHead(c) {
			headOnly(c, http.StatusOK, htmlContentType)
			return
		}
		c.HTML(http.StatusOK, "plain.tmpl", page.PlainView())
	}
}

func HandleRobots(page *render.Page) gin.HandlerFunc {
	return func(c *gin.Context) {
		const contentType = "text/plain; charset=utf-8"
		if isHead(c) {
			headOnly(c, http.StatusOK, contentType)
			return
		}
		c.Data(http.StatusOK, contentType, []byte(page.Robots()))
	}
}

func HandleSitemap(page *render.Page) gin.HandlerFunc {
	return func(c *gin.Context) {
		const contentType = "application/xml; charset=utf-8"
		body, err := page.Sitemap()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		if isHead(c) {
			headOnly(c, http.StatusOK, contentType)
			return
		}
		c.Data(http.StatusOK, contentType, body)
	}
}

func isAPIPath(path string) bool {
	return path == "/api" || strings.HasPrefix(path, "/api/")
}

func HandleNotFound(page *render.Page) gin.HandlerFunc {
	index := HandleDefault(page)

	return func(c *gin.Context) {
		method := c.Request.Method
		api := isAPIPath(c.Request.URL.Path)
		if method != http.MethodGet && method != http.MethodHead {
			c.Header("Allow", allowedMethods)
			if api {
				responder.MethodNotAllowedResponse(c)
				return
			}
			c.Data(http.StatusMethodNotAllowed, "text/plain; charset=utf-8", []byte("method not allowed\n"))
			return
		}
		if api {
			if isHead(c) {
				headOnly(c, http.StatusNotFound, "application/json; charset=utf-8")
				return
			}
			responder.NotFoundResponse(c)
			return
		}
		index(c)
	}
}
