package routes

import (
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/messaging"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/render"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/camera"
)

type Deps struct {
	Page       *render.Page
	Camera     camera.Stream
	QueryBus   messaging.QueryBus
	CommandBus messaging.CommandBus
}

func Assemble(d Deps) Router {
	router := NewRouter()

	router.Config()
	router.Headers()
	router.Fileserver()
	router.Default(d.Page)
	router.Camera(d.Camera)
	router.Feeds(d.QueryBus, d.CommandBus)

	return router
}
