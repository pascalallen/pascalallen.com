package auth

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/domain/auth/user"
	"github.com/pascalallen/pascalallen.com/http"
	"github.com/pascalallen/pascalallen.com/service/tokenservice"
)

type RefreshTokensRules struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func HandleRefreshTokens(c *gin.Context, userRepository user.UserRepository) {
	var request RefreshTokensRules

	if err := c.ShouldBind(&request); err != nil {
		errorMessage := fmt.Sprintf("Request validation error: %s", err.Error())
		http.BadRequestResponse(c, errors.New(errorMessage))

		return
	}

	userClaims := tokenservice.ParseAccessToken(request.AccessToken)
	refreshClaims := tokenservice.ParseRefreshToken(request.RefreshToken)

	u, err := userRepository.GetById(ulid.MustParse(userClaims.Id))
	if u == nil || err != nil {
		errorMessage := "invalid credentials"
		http.UnauthorizedResponse(c, errors.New(errorMessage))

		return
	}

	// refresh token is expired
	if refreshClaims.Valid() != nil {
		request.RefreshToken, err = tokenservice.NewRefreshToken(*refreshClaims)
		if err != nil {
			errorMessage := "error creating refresh token"
			http.InternalServerErrorResponse(c, errors.New(errorMessage))

			return
		}
	}

	// access token is expired
	if userClaims.StandardClaims.Valid() != nil && refreshClaims.Valid() == nil {
		request.AccessToken, err = tokenservice.NewAccessToken(*userClaims)
		if err != nil {
			errorMessage := "error creating access token"
			http.InternalServerErrorResponse(c, errors.New(errorMessage))

			return
		}
	}

	var roles []string
	for _, r := range u.Roles {
		roles = append(roles, r.Name)
	}

	var permissions []string
	for _, p := range u.Permissions() {
		permissions = append(permissions, p.Name)
	}

	userData := UserData{
		Id:           ulid.ULID(u.Id).String(),
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		EmailAddress: u.EmailAddress,
		CreatedAt:    u.CreatedAt.String(),
		ModifiedAt:   u.ModifiedAt.String(),
	}
	if !u.DeletedAt.IsZero() {
		userData.DeletedAt = u.DeletedAt.String()
	}

	responseData := &ResponseData{
		AccessToken:  request.AccessToken,
		RefreshToken: request.RefreshToken,
		User:         userData,
		Roles:        roles,
		Permissions:  permissions,
	}

	http.CreatedResponse(c, responseData)

	return
}
