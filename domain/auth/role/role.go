package role

import (
	"github.com/oklog/ulid/v2"
	_type "github.com/pascalallen/pascalallen.com/database/type"
	"github.com/pascalallen/pascalallen.com/domain/auth/permission"
	"time"
)

type Role struct {
	Id          _type.GormUlid          `json:"id" gorm:"primaryKey;size:26;not null"`
	Name        string                  `json:"name"`
	Permissions []permission.Permission `json:"permissions,omitempty" gorm:"many2many:role_permissions"`
	CreatedAt   time.Time               `json:"created_at"`
	ModifiedAt  time.Time               `json:"modified_at"`
}

func Define(id ulid.ULID, name string) *Role {
	createdAt := time.Now()

	return &Role{
		Id:         _type.GormUlid(id),
		Name:       name,
		CreatedAt:  createdAt,
		ModifiedAt: createdAt,
	}
}

func (r *Role) UpdateName(name string) {
	r.Name = name
	r.ModifiedAt = time.Now()
}

func (r *Role) AddPermission(permission permission.Permission) {
	for _, p := range r.Permissions {
		if p.Id == permission.Id {
			return
		}
	}

	r.Permissions = append(r.Permissions, permission)
	r.ModifiedAt = time.Now()
}

func (r *Role) RemovePermission(permission permission.Permission) {
	for i, p := range r.Permissions {
		if p.Id == permission.Id {
			r.Permissions[i] = r.Permissions[len(r.Permissions)-1]
		}
	}

	r.Permissions = r.Permissions[:len(r.Permissions)-1]
}

func (r *Role) HasPermission(name string) bool {
	for _, p := range r.Permissions {
		if p.Name == name {
			return true
		}
	}

	return false
}
