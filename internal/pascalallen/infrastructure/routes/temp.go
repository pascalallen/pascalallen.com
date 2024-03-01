package routes

import (
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/action"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/middleware"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/user"
)

func (r Router) Temp(repository user.UserRepository) {
	v := r.engine.Group(v1)
	{
		v.GET(
			"/temp",
			middleware.AuthRequired(repository),
			action.HandleTemp(),
		)

		ch := make(chan string)
		v.POST(
			"/event-stream",
			middleware.AuthRequired(repository),
			action.HandleEventStreamPost(ch),
		)
		v.GET(
			"/event-stream",
			middleware.EventStreamHeaders(),
			action.HandleEventStreamGet(ch),
		)
	}
}
