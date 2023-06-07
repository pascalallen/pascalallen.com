package auth

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/domain/auth/user"
	"github.com/pascalallen/pascalallen.com/http"
	"github.com/pascalallen/pascalallen.com/service"
	"time"
)

type LoginUserRules struct {
	EmailAddress string `form:"email_address" json:"email_address" binding:"required,max=100,email"`
	Password     string `form:"password" json:"password" binding:"required"`
}

type UserData struct {
	Id           string `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	EmailAddress string `json:"email_address"`
	CreatedAt    string `json:"created_at"`
	ModifiedAt   string `json:"modified_at"`
	DeletedAt    string `json:"deleted_at,omitempty"`
}

type ResponseData struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	User         UserData `json:"user"`
	Roles        []string `json:"roles,omitempty"`
	Permissions  []string `json:"permissions,omitempty"`
}

func HandleLoginUser(c *gin.Context, userRepository user.UserRepository) {
	var request LoginUserRules

	if err := c.ShouldBind(&request); err != nil {
		errorMessage := fmt.Sprintf("Request validation error: %s", err.Error())
		http.BadRequestResponse(c, errors.New(errorMessage))

		return
	}

	u, err := userRepository.GetByEmailAddress(request.EmailAddress)
	if u == nil || err != nil {
		errorMessage := "invalid credentials"
		http.UnauthorizedResponse(c, errors.New(errorMessage))

		return
	}

	if u.PasswordHash.Compare(request.Password) == false {
		errorMessage := "invalid credentials"
		http.UnauthorizedResponse(c, errors.New(errorMessage))

		return
	}

	userClaims := service.UserClaims{
		Id:    ulid.ULID(u.Id).String(),
		First: u.FirstName,
		Last:  u.LastName,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
	}

	signedAccessToken, err := service.NewAccessToken(userClaims)
	if err != nil {
		errorMessage := "error creating access token"
		http.InternalServerErrorResponse(c, errors.New(errorMessage))

		return
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

	refreshClaims := jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 48).Unix(),
	}

	signedRefreshToken, err := service.NewRefreshToken(refreshClaims)
	if err != nil {
		errorMessage := "error creating refresh token"
		http.InternalServerErrorResponse(c, errors.New(errorMessage))

		return
	}

	responseData := &ResponseData{
		AccessToken:  signedAccessToken,
		RefreshToken: signedRefreshToken,
		User:         userData,
		Roles:        roles,
		Permissions:  permissions,
	}

	http.CreatedResponse(c, responseData)

	return
}
