package routes

import (
	"github.com/gin-gonic/gin"
	"log"
)

func Config(router *gin.Engine) {
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatal(err)
	}

	router.LoadHTMLGlob("templates/*")
}
