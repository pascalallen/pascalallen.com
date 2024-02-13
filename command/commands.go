package command

import "github.com/oklog/ulid/v2"

type RegisterUser struct {
	Id           ulid.ULID
	FirstName    string
	LastName     string
	EmailAddress string
}

type UpdateUser struct {
	Id           ulid.ULID
	FirstName    string
	LastName     string
	EmailAddress string
}
