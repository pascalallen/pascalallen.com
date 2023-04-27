package main

import (
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

var env = map[string]string{
	"APP_BASE_URL": os.Getenv("APP_BASE_URL"),
}

func main() {
	envBytes, _ := json.Marshal(env)

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/public", "./public")
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"environment": base64.StdEncoding.EncodeToString(envBytes),
		})
	})

	log.Fatalf("error running HTTP server: %s\n", router.Run(":9990"))
}
