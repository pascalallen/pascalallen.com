package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pascalallen/pascalallen.com/domain/user"
	"github.com/pascalallen/pascalallen.com/http/action"
	"github.com/pascalallen/pascalallen.com/http/middleware"
)

func Temp(router *gin.Engine, userRepository user.UserRepository) {
	router.GET(
		"/api/v1/temp",
		middleware.AuthRequired(userRepository),
		action.HandleTemp(),
	)
	ch := make(chan string)
	router.POST(
		"/api/v1/event-stream",
		middleware.AuthRequired(userRepository),
		action.HandleEventStreamPost(ch),
	)
	router.GET(
		"/api/v1/event-stream",
		middleware.EventStreamHeaders(),
		action.HandleEventStreamGet(ch),
	)
}
