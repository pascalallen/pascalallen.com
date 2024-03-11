package repository

import (
	"errors"
	"fmt"
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/permission"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/database"
	"gorm.io/gorm"
)

type GormPermissionRepository struct {
	session database.Session
}

func NewGormPermissionRepository(session database.Session) *GormPermissionRepository {
	return &GormPermissionRepository{
		session: session,
	}
}

func (repository *GormPermissionRepository) GetById(id ulid.ULID) (*permission.Permission, error) {
	var p *permission.Permission
	if err := repository.session.First(&p, "id = ?", id.String()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to fetch Permission by ID: %s", id)
	}

	return p, nil
}

func (repository *GormPermissionRepository) GetByName(name string) (*permission.Permission, error) {
	var p *permission.Permission
	if err := repository.session.First(&p, "name = ?", name); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to fetch Permission by name: %s", name)
	}

	return p, nil
}

// GetAll TODO: Add pagination
func (repository *GormPermissionRepository) GetAll() (*[]permission.Permission, error) {
	var permissions *[]permission.Permission
	if err := repository.session.Find(&permissions); err != nil {
		return nil, fmt.Errorf("failed to fetch all Permissions: %s", err)
	}

	return permissions, nil
}

func (repository *GormPermissionRepository) Add(permission *permission.Permission) error {
	if err := repository.session.Create(&permission); err != nil {
		return fmt.Errorf("failed to persist Permission to database: %s", permission)
	}

	return nil
}

func (repository *GormPermissionRepository) Remove(permission *permission.Permission) error {
	if err := repository.session.Delete(&permission); err != nil {
		return fmt.Errorf("failed to delete Permission from database: %s", permission)
	}

	return nil
}

func (repository *GormPermissionRepository) UpdateOrAdd(permission *permission.Permission) error {
	if err := repository.session.Save(&permission); err != nil {
		return fmt.Errorf("failed to update Permission: %s", permission)
	}

	return nil
}
