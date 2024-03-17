// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/database"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/repository"
)

import (
	_ "github.com/joho/godotenv/autoload"
)

// Injectors from wire.go:

func InitializeContainer() Container {
	session := database.NewGormSession()
	permissionRepository := repository.NewGormPermissionRepository(session)
	roleRepository := repository.NewGormRoleRepository(session)
	userRepository := repository.NewGormUserRepository(session)
	seeder := database.NewDatabaseSeeder(session, permissionRepository, roleRepository, userRepository)
	container := NewContainer(session, permissionRepository, roleRepository, userRepository, seeder)
	return container
}
