package repository

import (
	"errors"
	"fmt"
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/domain/permission"
	"gorm.io/gorm"
)

type GormPermissionRepository struct {
	unitOfWork *gorm.DB
}

func NewGormPermissionRepository(unitOfWork *gorm.DB) GormPermissionRepository {
	return GormPermissionRepository{
		unitOfWork: unitOfWork,
	}
}

func (repository GormPermissionRepository) GetById(id ulid.ULID) (*permission.Permission, error) {
	var p *permission.Permission
	if err := repository.unitOfWork.First(&p, "id = ?", id.String()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to fetch Permission by ID: %s", id)
	}

	return p, nil
}

func (repository GormPermissionRepository) GetByName(name string) (*permission.Permission, error) {
	var p *permission.Permission
	if err := repository.unitOfWork.First(&p, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to fetch Permission by name: %s", name)
	}

	return p, nil
}

// GetAll TODO: Add pagination
func (repository GormPermissionRepository) GetAll() (*[]permission.Permission, error) {
	var permissions *[]permission.Permission
	if err := repository.unitOfWork.Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch all Permissions: %s", err)
	}

	return permissions, nil
}

func (repository GormPermissionRepository) Add(permission *permission.Permission) error {
	if err := repository.unitOfWork.Create(&permission).Error; err != nil {
		return fmt.Errorf("failed to persist Permission to database: %s", permission)
	}

	return nil
}

func (repository GormPermissionRepository) Remove(permission *permission.Permission) error {
	if err := repository.unitOfWork.Delete(&permission).Error; err != nil {
		return fmt.Errorf("failed to delete Permission from database: %s", permission)
	}

	return nil
}

func (repository GormPermissionRepository) UpdateOrAdd(permission *permission.Permission) error {
	if err := repository.unitOfWork.Save(&permission).Error; err != nil {
		return fmt.Errorf("failed to update Permission: %s", permission)
	}

	return nil
}
