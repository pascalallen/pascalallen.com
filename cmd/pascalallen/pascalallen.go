package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/command"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/command_handler"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/query"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/query_handler"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/render"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/content"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/camera"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/container"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/routes"
	"os"
	"time"
)

func main() {
	gin.SetMode(os.Getenv("GIN_MODE"))

	c := container.InitializeContainer()

	c.QueryBus.RegisterHandler((&query.ListRepositories{}).QueryName(), &query_handler.ListRepositoriesHandler{
		Feed: c.RepositoryFeed,
	})
	c.QueryBus.RegisterHandler((&query.ListPackages{}).QueryName(), &query_handler.ListPackagesHandler{
		Feed: c.PackageFeed,
	})
	c.CommandBus.RegisterHandler((&command.RecordVisit{}).CommandName(), &command_handler.RecordVisitHandler{
		Notifier: c.VisitNotifier,
		Now:      time.Now,
	})

	router := routes.Assemble(routes.Deps{
		Page: render.NewPage(content.FS(), render.Config{BaseURL: os.Getenv("APP_BASE_URL"), Env: os.Getenv("APP_ENV")}),
		Camera: camera.NewHub(camera.Config{
			Source: camera.GStreamerSource,
			Check:  camera.GStreamerAvailable,
		}),
		QueryBus:   c.QueryBus,
		CommandBus: c.CommandBus,
	})
	router.Serve(":" + port())
}

func port() string {
	if p := os.Getenv("PORT"); p != "" {
		return p
	}

	return "9990"
}
