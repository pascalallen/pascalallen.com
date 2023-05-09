package repository

import (
	"errors"
	"fmt"
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/domain/auth/role"
	"gorm.io/gorm"
)

type GormRoleRepository struct {
	UnitOfWork *gorm.DB
}

func (repository GormRoleRepository) GetById(id ulid.ULID) (*role.Role, error) {
	var r *role.Role
	if err := repository.UnitOfWork.Preload("Permissions").First(&r, "id = ?", id.String()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to fetch Role by ID: %s", id)
	}

	return r, nil
}

func (repository GormRoleRepository) GetByName(name string) (*role.Role, error) {
	var r *role.Role
	if err := repository.UnitOfWork.Preload("Permissions").First(&r, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to fetch Role by name: %s", name)
	}

	return r, nil
}

// GetAll TODO: Add pagination
func (repository GormRoleRepository) GetAll() (*[]role.Role, error) {
	var roles *[]role.Role
	if err := repository.UnitOfWork.Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch all Roles: %s", err)
	}

	return roles, nil
}

func (repository GormRoleRepository) Add(role *role.Role) error {
	if err := repository.UnitOfWork.Create(&role).Error; err != nil {
		return fmt.Errorf("failed to persist Role to database: %s", role)
	}

	return nil
}

func (repository GormRoleRepository) Remove(role *role.Role) error {
	if err := repository.UnitOfWork.Delete(&role).Error; err != nil {
		return fmt.Errorf("failed to delete Role from database: %s", role)
	}

	return nil
}

func (repository GormRoleRepository) UpdateOrAdd(role *role.Role) error {
	if err := repository.UnitOfWork.Save(&role).Error; err != nil {
		return fmt.Errorf("failed to update Role: %s", role)
	}

	return nil
}
