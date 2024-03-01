package role

import (
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/permission"
	"testing"
)

func TestThatDefineReturnsInstanceOfRole(t *testing.T) {
	id := ulid.Make()
	name := "ROLE_USER"
	r := Define(id, name)

	if r == nil {
		t.Fatal("test failed attempting to call method: Define")
	}
}

func TestThatUpdateNameUpdatesName(t *testing.T) {
	id := ulid.Make()
	name := "ROLE_USER"
	r := Define(id, name)

	updatedName := "ROLE_SUPER_ADMIN"
	r.UpdateName(updatedName)

	if r.Name == name {
		t.Fatal("test failed attempting to call method: UpdateName")
	}
}

func TestThatAddPermissionAddsPermission(t *testing.T) {
	id := ulid.Make()
	name := "ROLE_SUPER_ADMIN"
	r := Define(id, name)
	permissionName := "CREATE_ROLES"
	p := permission.Define(ulid.Make(), permissionName, "Allows the user to create roles")
	r.AddPermission(*p)

	if r.HasPermission(permissionName) == false {
		t.Fatal("test failed attempting to call method: AddPermission")
	}
}

func TestThatAddPermissionReturnsEarlyIfPermissionAlreadyExists(t *testing.T) {
	id := ulid.Make()
	name := "ROLE_SUPER_ADMIN"
	r := Define(id, name)
	permissionName := "CREATE_ROLES"
	p := permission.Define(ulid.Make(), permissionName, "Allows the user to create roles")
	r.AddPermission(*p)
	r.AddPermission(*p)

	if r.HasPermission(permissionName) == false {
		t.Fatal("test failed attempting to call method: AddPermission")
	}
}

func TestThatRemovePermissionRemovesPermission(t *testing.T) {
	id := ulid.Make()
	name := "ROLE_SUPER_ADMIN"
	r := Define(id, name)
	permissionName := "CREATE_ROLES"
	p := permission.Define(ulid.Make(), permissionName, "Allows the user to create roles")
	r.AddPermission(*p)
	r.RemovePermission(*p)

	if r.HasPermission(permissionName) == true {
		t.Fatal("test failed attempting to call method: RemovePermission")
	}
}
