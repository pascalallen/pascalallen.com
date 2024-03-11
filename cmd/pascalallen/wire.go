//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/providers"
)

func InitializeContainer() *Container {
	wire.Build(NewContainer, providers.NewDatabaseSession, providers.NewPermissionRepository, providers.NewRoleRepository, providers.NewUserRepository, providers.NewDatabaseSeeder)
	return &Container{}
}
