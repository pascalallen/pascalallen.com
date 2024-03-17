//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/database"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/repository"
)

func InitializeContainer() Container {
	wire.Build(NewContainer, database.NewGormSession, repository.NewGormPermissionRepository, repository.NewGormRoleRepository, repository.NewGormUserRepository, database.NewDatabaseSeeder)
	return Container{}
}
