package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/camera"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/routes"
	"os"
)

func main() {
	gin.SetMode(os.Getenv("GIN_MODE"))

	router := routes.NewRouter()

	router.Config()
	router.Fileserver()
	router.Default()
	router.Camera(camera.NewHub(camera.Config{
		Source: camera.GStreamerSource,
		Check:  camera.GStreamerAvailable,
	}))
	router.Serve(":" + port())
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}

	return "9990"
}
