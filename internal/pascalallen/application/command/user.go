package command

import (
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/password"
	"reflect"
)

type RegisterUser struct {
	Id           ulid.ULID             `json:"id"`
	FirstName    string                `json:"first_name"`
	LastName     string                `json:"last_name"`
	EmailAddress string                `json:"email_address"`
	PasswordHash password.PasswordHash `json:"password_hash"`
}

func (c RegisterUser) CommandName() string {
	return reflect.TypeOf(c).Name()
}

type UpdateUser struct {
	Id           ulid.ULID
	FirstName    string
	LastName     string
	EmailAddress string
}

func (c UpdateUser) CommandName() string {
	return reflect.TypeOf(c).Name()
}
