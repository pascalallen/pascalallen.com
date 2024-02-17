package command

import (
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/domain/password"
	"reflect"
)

type Command interface {
	CommandName() string
}

type RegisterUser struct {
	Id           ulid.ULID
	FirstName    string
	LastName     string
	EmailAddress string
	PasswordHash password.PasswordHash
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
