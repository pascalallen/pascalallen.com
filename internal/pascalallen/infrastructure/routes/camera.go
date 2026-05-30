package routes

import "github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/action"

func (r Router) Camera() {
	v := r.engine.Group(v1)
	{
		camera := v.Group("/camera")
		{
			camera.GET("/stream", action.HandleCameraStream())
		}
	}
}
