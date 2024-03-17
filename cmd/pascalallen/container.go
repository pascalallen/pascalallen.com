package main

import (
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/permission"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/role"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/user"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/database"
)

type Container struct {
	DatabaseSession      database.Session
	PermissionRepository permission.PermissionRepository
	RoleRepository       role.RoleRepository
	UserRepository       user.UserRepository
	DatabaseSeeder       database.Seeder
}

func NewContainer(dbSession database.Session, permissionRepo permission.PermissionRepository, roleRepo role.RoleRepository, userRepo user.UserRepository, dbSeeder database.Seeder) Container {
	return Container{
		DatabaseSession:      dbSession,
		PermissionRepository: permissionRepo,
		RoleRepository:       roleRepo,
		UserRepository:       userRepo,
		DatabaseSeeder:       dbSeeder,
	}
}
