package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pascalallen/pascalallen.com/http/action"
)

func Default(router *gin.Engine) {
	router.NoRoute(action.HandleDefault())
}
