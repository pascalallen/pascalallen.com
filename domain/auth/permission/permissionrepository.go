package permission

import (
	"github.com/oklog/ulid/v2"
)

type PermissionRepository interface {
	GetById(id ulid.ULID) (*Permission, error)
	GetByName(name string) (*Permission, error)
	// TODO: Include pagination
	GetAll() (*[]Permission, error)
	Add(permission *Permission) error
	Remove(permission *Permission) error
	UpdateOrAdd(permission *Permission) error
}
