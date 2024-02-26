package routes

import (
	"github.com/pascalallen/pascalallen.com/domain/user"
	"github.com/pascalallen/pascalallen.com/http/action"
	"github.com/pascalallen/pascalallen.com/http/middleware"
)

func (r Router) Temp(repository user.UserRepository) {
	r.engine.GET(
		"/api/v1/temp",
		middleware.AuthRequired(repository),
		action.HandleTemp(),
	)
	ch := make(chan string)
	r.engine.POST(
		"/api/v1/event-stream",
		middleware.AuthRequired(repository),
		action.HandleEventStreamPost(ch),
	)
	r.engine.GET(
		"/api/v1/event-stream",
		middleware.EventStreamHeaders(),
		action.HandleEventStreamGet(ch),
	)
}
