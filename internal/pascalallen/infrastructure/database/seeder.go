package database

import (
	"encoding/json"
	"fmt"
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/permission"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/role"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/user"
	"log"
	"os"
	"path"
	"runtime"
)

type PermissionData struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PermissionsData struct {
	Permissions []PermissionData `json:"permissions"`
}

type RoleData struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type RolesData struct {
	Roles []RoleData `json:"roles"`
}

type UserData struct {
	Id           string   `json:"id"`
	FirstName    string   `json:"first_name"`
	LastName     string   `json:"last_name"`
	EmailAddress string   `json:"email_address"`
	Roles        []string `json:"roles"`
}

type UsersData struct {
	Users []UserData `json:"users"`
}

type Seeder interface {
	Seed()
}

type DatabaseSeeder struct {
	permissionsMap       map[string]permission.Permission
	rolesMap             map[string]role.Role
	session              Session
	permissionRepository permission.PermissionRepository
	roleRepository       role.RoleRepository
	userRepository       user.UserRepository
}

func NewDatabaseSeeder(session Session, permissionRepo permission.PermissionRepository, roleRepo role.RoleRepository, userRepo user.UserRepository) Seeder {
	return &DatabaseSeeder{
		permissionsMap:       make(map[string]permission.Permission),
		rolesMap:             make(map[string]role.Role),
		session:              session,
		permissionRepository: permissionRepo,
		roleRepository:       roleRepo,
		userRepository:       userRepo,
	}
}

func (s *DatabaseSeeder) Seed() {
	if err := s.seedPermissions(); err != nil {
		log.Fatalf("failed to seed permissions: %s", err)
	}

	if err := s.seedRoles(); err != nil {
		log.Fatalf("failed to seed roles: %s", err)
	}

	if err := s.seedUsers(); err != nil {
		log.Fatalf("failed to seed users: %s", err)
	}
}

func (s *DatabaseSeeder) seedPermissions() error {
	if err := s.loadPermissionsMap(); err != nil {
		return fmt.Errorf("failed to load permissions map: %s", err)
	}

	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("error getting filename")
	}

	rootDir := path.Dir(filename)
	permissionsFile := fmt.Sprintf("%s/seeds/auth.permissions.json", rootDir)

	contents, err := os.ReadFile(permissionsFile)
	if err != nil {
		return fmt.Errorf("error reading permissions file: %s", err)
	}

	var permissionsData PermissionsData
	if err := json.Unmarshal(contents, &permissionsData); err != nil {
		return fmt.Errorf("failed to parse Permission seed file contents: %s, error: %s", contents, err)
	}

	var currentPermissions []string
	for permissionName := range s.permissionsMap {
		currentPermissions = append(currentPermissions, permissionName)
	}

	var seedPermissions []string
	for _, permissionData := range permissionsData.Permissions {
		seedPermissions = append(seedPermissions, permissionData.Name)
	}

	var permissionsToRemove []string
	for _, permissionName := range seedPermissions {
		if len(currentPermissions) > 0 && !contains(currentPermissions, permissionName) {
			permissionsToRemove = append(permissionsToRemove, permissionName)
		}
	}

	for _, permissionName := range permissionsToRemove {
		p := s.permissionsMap[permissionName]
		if err := s.permissionRepository.Remove(&p); err != nil {
			return err
		}
	}

	for _, permissionData := range permissionsData.Permissions {
		id := ulid.MustParse(permissionData.Id)

		p, err := s.permissionRepository.GetById(id)
		if err != nil {
			return err
		}

		if p == nil {
			p = permission.Define(id, permissionData.Name, permissionData.Description)
			if err := s.permissionRepository.Add(p); err != nil {
				return err
			}
		}

		if permissionData.Name != p.Name {
			p.UpdateName(permissionData.Name)
		}

		if permissionData.Description != p.Description {
			p.UpdateDescription(permissionData.Description)
		}

		if err := s.permissionRepository.UpdateOrAdd(p); err != nil {
			return err
		}
	}

	if err := s.loadPermissionsMap(); err != nil {
		return fmt.Errorf("failed to load permissions map: %s", err)
	}

	return nil
}

func (s *DatabaseSeeder) seedRoles() error {
	if err := s.loadRolesMap(); err != nil {
		return fmt.Errorf("failed to load roles map: %s", err)
	}

	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("error getting filename")
	}

	rootDir := path.Dir(filename)
	rolesFile := fmt.Sprintf("%s/seeds/auth.roles.json", rootDir)

	contents, err := os.ReadFile(rolesFile)
	if err != nil {
		return fmt.Errorf("error reading roles file: %s", err)
	}

	var rolesData RolesData
	if err := json.Unmarshal(contents, &rolesData); err != nil {
		return fmt.Errorf("failed to parse Role seed file contents: %s, error: %s", contents, err)
	}

	var currentRoles []string
	for roleName := range s.rolesMap {
		currentRoles = append(currentRoles, roleName)
	}

	var seedRoles []string
	for _, roleData := range rolesData.Roles {
		seedRoles = append(seedRoles, roleData.Name)
	}

	var rolesToRemove []string
	for _, roleName := range seedRoles {
		if len(currentRoles) > 0 && !contains(currentRoles, roleName) {
			rolesToRemove = append(rolesToRemove, roleName)
		}
	}

	for _, roleName := range rolesToRemove {
		r := s.rolesMap[roleName]
		if err := s.roleRepository.Remove(&r); err != nil {
			return err
		}
	}

	for _, roleData := range rolesData.Roles {
		id := ulid.MustParse(roleData.Id)

		r, err := s.roleRepository.GetById(id)
		if err != nil {
			return err
		}

		if r == nil {
			r = role.Define(id, roleData.Name)
			if len(roleData.Permissions) > 0 {
				for _, permissionName := range roleData.Permissions {
					p, err := s.permissionRepository.GetByName(permissionName)
					if err != nil {
						return err
					}

					if p != nil && !r.HasPermission(permissionName) {
						r.AddPermission(*p)
					}
				}
			}

			if err := s.roleRepository.Add(r); err != nil {
				return err
			}
		}

		if roleData.Name != r.Name {
			r.UpdateName(roleData.Name)
		}

		var newRolePermissions []permission.Permission
		for _, permissionName := range roleData.Permissions {
			p, err := s.permissionRepository.GetByName(permissionName)
			if err != nil {
				return err
			}

			newRolePermissions = append(newRolePermissions, *p)
		}

		if err := s.session.Replace(r, "Permissions", newRolePermissions); err != nil {
			return fmt.Errorf("failed to update Role permissions: %s, error: %s", newRolePermissions, err)
		}

		if err := s.roleRepository.UpdateOrAdd(r); err != nil {
			return err
		}
	}

	if err := s.loadRolesMap(); err != nil {
		return fmt.Errorf("failed to load roles map: %s", err)
	}

	return nil
}

func (s *DatabaseSeeder) seedUsers() error {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("error getting filename")
	}

	rootDir := path.Dir(filename)
	usersFile := fmt.Sprintf("%s/seeds/auth.users.json", rootDir)

	contents, err := os.ReadFile(usersFile)
	if err != nil {
		return fmt.Errorf("error reading users file: %s", err)
	}

	var usersData UsersData
	if err := json.Unmarshal(contents, &usersData); err != nil {
		return fmt.Errorf("failed to parse User seed file contents: %s, error: %s", contents, err)
	}

	for _, userData := range usersData.Users {
		u, err := s.userRepository.GetByEmailAddress(userData.EmailAddress)
		if err != nil {
			return err
		}

		if u == nil {
			id := ulid.MustParse(userData.Id)
			u = user.Register(
				id,
				userData.FirstName,
				userData.LastName,
				userData.EmailAddress,
			)
			if len(userData.Roles) > 0 {
				for _, roleName := range userData.Roles {
					r, err := s.roleRepository.GetByName(roleName)
					if err != nil {
						return err
					}

					if r != nil && !u.HasRole(roleName) {
						u.AddRole(*r)
					}
				}
			}

			if err := s.userRepository.Add(u); err != nil {
				return err
			}
		}

		if userData.FirstName != u.FirstName {
			u.UpdateFirstName(userData.FirstName)
		}

		if userData.LastName != u.LastName {
			u.UpdateLastName(userData.LastName)
		}

		if userData.EmailAddress != u.EmailAddress {
			u.UpdateEmailAddress(userData.EmailAddress)
		}

		var newUserRoles []role.Role
		for _, roleName := range userData.Roles {
			r, err := s.roleRepository.GetByName(roleName)
			if err != nil {
				return err
			}

			newUserRoles = append(newUserRoles, *r)
		}

		if err := s.session.Replace(&u, "Roles", newUserRoles); err != nil {
			return fmt.Errorf("failed to update User roles: %s, error: %s", newUserRoles, err)
		}

		if err := s.userRepository.UpdateOrAdd(u); err != nil {
			return err
		}
	}

	return nil
}

func (s *DatabaseSeeder) loadPermissionsMap() error {
	permissions, err := s.permissionRepository.GetAll()
	if err != nil {
		return err
	}

	for _, p := range *permissions {
		s.permissionsMap[p.Name] = p
	}

	return nil
}

func (s *DatabaseSeeder) loadRolesMap() error {
	roles, err := s.roleRepository.GetAll()
	if err != nil {
		return err
	}

	for _, r := range *roles {
		s.rolesMap[r.Name] = r
	}

	return nil
}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}
