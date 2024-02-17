package database

import (
	"fmt"
	"github.com/pascalallen/pascalallen.com/domain/permission"
	"github.com/pascalallen/pascalallen.com/domain/role"
	"github.com/pascalallen/pascalallen.com/domain/user"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

func NewGormUnitOfWork() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %s", err)
	}

	return db, nil
}

func Migrate(unitOfWork *gorm.DB) {
	if err := unitOfWork.AutoMigrate(&permission.Permission{}, &role.Role{}, &user.User{}); err != nil {
		err := fmt.Errorf("failed to auto migrate database: %s", err)
		log.Fatal(err)
	}
}

func Seed(unitOfWork *gorm.DB, permissionRepository permission.PermissionRepository, roleRepository role.RoleRepository, userRepository user.UserRepository) {
	dataSeeder := DataSeeder{
		permissionsMap:       make(map[string]permission.Permission),
		rolesMap:             make(map[string]role.Role),
		UnitOfWork:           unitOfWork,
		PermissionRepository: permissionRepository,
		RoleRepository:       roleRepository,
		UserRepository:       userRepository,
	}
	if err := dataSeeder.Seed(); err != nil {
		log.Fatal(err)
	}
}
