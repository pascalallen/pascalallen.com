package auth

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/command"
	"github.com/pascalallen/pascalallen.com/domain/auth/passwordhash"
	"github.com/pascalallen/pascalallen.com/domain/auth/user"
	"github.com/pascalallen/pascalallen.com/http"
	"github.com/pascalallen/pascalallen.com/messaging"
)

type RegisterUserRules struct {
	FirstName       string `form:"first_name" json:"first_name" binding:"required,max=100"`
	LastName        string `form:"last_name" json:"last_name" binding:"required,max=100"`
	EmailAddress    string `form:"email_address" json:"email_address" binding:"required,max=100,email"`
	Password        string `form:"password" json:"password" binding:"required"`
	ConfirmPassword string `form:"confirm_password" json:"confirm_password" binding:"required,eqfield=Password"`
}

func HandleRegisterUser(userRepository user.UserRepository, commandBus messaging.CommandBus) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request RegisterUserRules

		if err := c.ShouldBind(&request); err != nil {
			errorMessage := fmt.Sprintf("Request validation error: %s", err.Error())
			http.BadRequestResponse(c, errors.New(errorMessage))

			return
		}

		if u, err := userRepository.GetByEmailAddress(request.EmailAddress); u != nil || err != nil {
			errorMessage := fmt.Sprint("Something went wrong. If you already have an account, please log in.")
			http.UnprocessableEntityResponse(c, errors.New(errorMessage))

			return
		}

		cmd := command.RegisterUser{
			Id:           ulid.Make(),
			FirstName:    request.FirstName,
			LastName:     request.LastName,
			EmailAddress: request.EmailAddress,
			PasswordHash: passwordhash.Create(request.Password),
		}
		commandBus.Execute(cmd)

		http.CreatedResponse[RegisterUserRules](c, &request)

		return
	}
}
