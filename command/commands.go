package command

import "github.com/oklog/ulid/v2"

type Command interface {
	GetName() string
}

type RegisterUser struct {
	Id           ulid.ULID
	FirstName    string
	LastName     string
	EmailAddress string
}

func (c RegisterUser) GetName() string {
	return "command.RegisterUser"
}

type UpdateUser struct {
	Id           ulid.ULID
	FirstName    string
	LastName     string
	EmailAddress string
}

func (c UpdateUser) GetName() string {
	return "command.UpdateUser"
}
