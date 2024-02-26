package routes

import (
	"github.com/pascalallen/pascalallen.com/domain/user"
	"github.com/pascalallen/pascalallen.com/http/action"
	"github.com/pascalallen/pascalallen.com/messaging"
)

func (r Router) Auth(repository user.UserRepository, bus messaging.CommandBus) {
	r.engine.POST("/api/v1/auth/register", action.HandleRegisterUser(repository, bus))
	r.engine.POST("/api/v1/auth/login", action.HandleLoginUser(repository))
	r.engine.PATCH("/api/v1/auth/refresh", action.HandleRefreshTokens(repository))
	// router.POST("/api/v1/auth/request-reset", auth.HandleRequestPasswordReset)
	// router.POST("/api/v1/auth/reset-password", auth.HandleResetPassword)
}
