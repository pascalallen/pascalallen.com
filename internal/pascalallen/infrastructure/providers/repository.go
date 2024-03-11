package providers

import (
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/permission"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/role"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/user"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/database"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/repository"
)

func NewPermissionRepository(session database.Session) permission.PermissionRepository {
	return repository.NewGormPermissionRepository(session)
}

func NewRoleRepository(session database.Session) role.RoleRepository {
	return repository.NewGormRoleRepository(session)
}

func NewUserRepository(session database.Session) user.UserRepository {
	return repository.NewGormUserRepository(session)
}
