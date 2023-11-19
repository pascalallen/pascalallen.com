package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/domain/auth/user"
	"github.com/pascalallen/pascalallen.com/http"
	"github.com/pascalallen/pascalallen.com/service/tokenservice"
	"strings"
)

func AuthRequired(userRepository user.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			http.BadRequestResponse(c, errors.New("authorization header is required"))

			return
		}

		accessToken := strings.Split(authHeader, " ")[1]
		userClaims := tokenservice.ParseAccessToken(accessToken)

		u, err := userRepository.GetById(ulid.MustParse(userClaims.Id))
		if u == nil || err != nil {
			http.UnauthorizedResponse(c, errors.New("invalid credentials"))

			return
		}

		c.Next()
	}
}
