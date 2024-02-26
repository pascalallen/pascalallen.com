package routes

import (
	"github.com/gin-gonic/gin"
	"log"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter() Router {
	return Router{
		engine: gin.Default(),
	}
}

func (r Router) Serve(port string) {
	if err := r.engine.Run(port); err != nil {
		log.Fatal(err)
	}
}
