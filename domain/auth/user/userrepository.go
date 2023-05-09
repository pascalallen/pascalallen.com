package user

import (
	"github.com/oklog/ulid/v2"
)

type UserRepository interface {
	GetById(id ulid.ULID) (*User, error)
	GetByEmailAddress(emailAddress string) (*User, error)
	// TODO: Include pagination
	GetAll(includeDeleted bool) (*[]User, error)
	Add(user *User) error
	Remove(user *User) error
	UpdateOrAdd(user *User) error
}
