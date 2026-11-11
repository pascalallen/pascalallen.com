package routes

import (
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/action"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/camera"
)

func (r Router) Camera(stream camera.Stream) {
	v := r.engine.Group(v1)
	{
		camera := v.Group("/camera")
		{
			camera.GET("/stream", action.HandleCameraStream(stream))
			camera.GET("/status", action.HandleCameraStatus(stream))
			camera.HEAD("/status", action.HandleCameraStatus(stream))
		}
	}
}
