package routes

import "github.com/gin-gonic/gin"

func Fileserver(router *gin.Engine) {
	router.Static("/public", "../public")
}
