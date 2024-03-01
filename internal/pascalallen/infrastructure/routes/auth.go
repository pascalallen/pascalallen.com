package routes

import (
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/http/action"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/user"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/messaging"
)

func (r Router) Auth(repository user.UserRepository, bus messaging.CommandBus) {
	v := r.engine.Group(v1)
	{
		auth := v.Group("/auth")
		{
			auth.POST("/register", action.HandleRegisterUser(repository, bus))
			auth.POST("/login", action.HandleLoginUser(repository))
			auth.PATCH("/refresh", action.HandleRefreshTokens(repository))
			// router.POST("/request-reset", auth.HandleRequestPasswordReset)
			// router.POST("/reset-password", auth.HandleResetPassword)
		}
	}
}
