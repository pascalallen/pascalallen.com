package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/pascalallen/pascalallen.com/domain/user"
	"github.com/pascalallen/pascalallen.com/http/action"
	"github.com/pascalallen/pascalallen.com/messaging"
)

func Auth(router *gin.Engine, userRepository user.UserRepository, commandBus messaging.CommandBus) {
	router.POST("/api/v1/auth/register", action.HandleRegisterUser(userRepository, commandBus))
	router.POST("/api/v1/auth/login", action.HandleLoginUser(userRepository))
	router.PATCH("/api/v1/auth/refresh", action.HandleRefreshTokens(userRepository))
	// router.POST("/api/v1/auth/request-reset", auth.HandleRequestPasswordReset)
	// router.POST("/api/v1/auth/reset-password", auth.HandleResetPassword)
}
