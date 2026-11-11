package routes

import (
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/action"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/render"
)

func (r Router) Default(page *render.Page) {
	r.engine.GET("/", action.HandleDefault(page))
	r.engine.HEAD("/", action.HandleDefault(page))
	r.engine.GET("/plain", action.HandlePlain(page))
	r.engine.HEAD("/plain", action.HandlePlain(page))
	r.engine.GET("/robots.txt", action.HandleRobots(page))
	r.engine.HEAD("/robots.txt", action.HandleRobots(page))
	r.engine.GET("/sitemap.xml", action.HandleSitemap(page))
	r.engine.HEAD("/sitemap.xml", action.HandleSitemap(page))
	r.engine.NoRoute(action.HandleNotFound(page))
}
