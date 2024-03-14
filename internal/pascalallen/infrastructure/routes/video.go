package routes

import "github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/action"

func (r Router) VideoFeed() {
	ch := make(chan []byte)
	r.engine.POST("/live", action.HandleVideoStreamPost(ch))
	r.engine.GET("/live", action.HandleVideoStreamGet(ch))
}
