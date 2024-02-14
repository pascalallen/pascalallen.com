package command

import "github.com/oklog/ulid/v2"

type Command interface {
	CommandName() string
}

type RegisterUser struct {
	Id           ulid.ULID
	FirstName    string
	LastName     string
	EmailAddress string
}

func (c RegisterUser) CommandName() string {
	return "command.RegisterUser"
}

type UpdateUser struct {
	Id           ulid.ULID
	FirstName    string
	LastName     string
	EmailAddress string
}

func (c UpdateUser) CommandName() string {
	return "command.UpdateUser"
}
