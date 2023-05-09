package role

import "github.com/oklog/ulid/v2"

type RoleRepository interface {
	GetById(id ulid.ULID) (*Role, error)
	GetByName(name string) (*Role, error)
	// TODO: Include pagination
	GetAll() (*[]Role, error)
	Add(role *Role) error
	Remove(role *Role) error
	UpdateOrAdd(role *Role) error
}
