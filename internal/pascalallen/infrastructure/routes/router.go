package routes

import (
	"github.com/gin-gonic/gin"
	"log"
)

const v1 = "/api/v1"

type Router struct {
	engine *gin.Engine
}

func NewRouter() Router {
	engine := gin.Default()
	engine.RedirectTrailingSlash = false
	engine.RedirectFixedPath = false

	return Router{engine: engine}
}

func (r Router) Serve(port string) {
	if err := r.engine.Run(port); err != nil {
		log.Fatal(err)
	}
}
