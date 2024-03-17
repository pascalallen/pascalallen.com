package repository

import (
	"errors"
	"fmt"
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/user"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/database"
	"gorm.io/gorm"
)

type GormUserRepository struct {
	session database.Session
}

func NewGormUserRepository(session database.Session) user.UserRepository {
	return &GormUserRepository{
		session: session,
	}
}

func (repository *GormUserRepository) GetById(id ulid.ULID) (*user.User, error) {
	var u *user.User
	if err := repository.session.Preload("Roles.Permissions").First(&u, "id = ?", id.String()); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to fetch User by ID: %s", id)
	}

	return u, nil
}

func (repository *GormUserRepository) GetByEmailAddress(emailAddress string) (*user.User, error) {
	var u *user.User
	if err := repository.session.Preload("Roles.Permissions").First(&u, "email_address = ?", emailAddress); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to fetch User by email address: %s", emailAddress)
	}

	return u, nil
}

// GetAll TODO: Add pagination
func (repository *GormUserRepository) GetAll(includeDeleted bool) (*[]user.User, error) {
	var users *[]user.User
	if !includeDeleted {
		repository.session = repository.session.Where("deleted_at IS NULL")
	}

	if err := repository.session.Find(&users); err != nil {
		return nil, fmt.Errorf("failed to fetch all Users: %s", err)
	}

	return users, nil
}

func (repository *GormUserRepository) Add(user *user.User) error {
	if err := repository.session.Create(user); err != nil {
		return fmt.Errorf("failed to persist User to database: %s", user)
	}

	return nil
}

func (repository *GormUserRepository) Remove(user *user.User) error {
	user.Delete()

	if err := repository.session.Save(&user); err != nil {
		return fmt.Errorf("failed to delete User from database: %s", user)
	}

	return nil
}

func (repository *GormUserRepository) UpdateOrAdd(user *user.User) error {
	if err := repository.session.Save(&user); err != nil {
		return fmt.Errorf("failed to update User: %s", user)
	}

	return nil
}
