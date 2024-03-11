package providers

import (
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/permission"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/role"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/user"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/database"
)

func NewDatabaseSession() database.Session {
	return database.NewGormSession()
}

func NewDatabaseSeeder(session database.Session, permissionRepo permission.PermissionRepository, roleRepo role.RoleRepository, userRepo user.UserRepository) database.Seeder {
	return database.NewDataSeeder(session, permissionRepo, roleRepo, userRepo)
}
