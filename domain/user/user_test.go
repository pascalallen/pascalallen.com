package user

import (
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/domain/password"
	"github.com/pascalallen/pascalallen.com/domain/permission"
	"github.com/pascalallen/pascalallen.com/domain/role"
	"testing"
)

func TestThatRegisterReturnsInstanceOfUser(t *testing.T) {
	id := ulid.Make()
	firstName := "Leeroy"
	lastName := "Jenkins"
	emailAddress := "ljenkins@example.com"
	u := Register(id, firstName, lastName, emailAddress)

	if u == nil {
		t.Fatal("test failed attempting to call method: Register")
	}
}

func TestThatUpdateFirstNameUpdatesFirstName(t *testing.T) {
	id := ulid.Make()
	firstName := "Leeroy"
	lastName := "Jenkins"
	emailAddress := "ljenkins@example.com"
	u := Register(id, firstName, lastName, emailAddress)

	updatedFirstName := "Bill"
	u.UpdateFirstName(updatedFirstName)

	if u.FirstName == firstName {
		t.Fatal("test failed attempting to call method: UpdateFirstName")
	}
}

func TestThatUpdateLastNameUpdatesLastName(t *testing.T) {
	id := ulid.Make()
	firstName := "Leeroy"
	lastName := "Jenkins"
	emailAddress := "ljenkins@example.com"
	u := Register(id, firstName, lastName, emailAddress)

	updatedLastName := "Cosby"
	u.UpdateLastName(updatedLastName)

	if u.LastName == lastName {
		t.Fatal("test failed attempting to call method: UpdateLastName")
	}
}

func TestThatUpdateEmailAddressUpdatesEmailAddress(t *testing.T) {
	id := ulid.Make()
	firstName := "Leeroy"
	lastName := "Jenkins"
	emailAddress := "ljenkins@example.com"
	u := Register(id, firstName, lastName, emailAddress)

	updatedEmailAddress := "bcosby@example.com"
	u.UpdateEmailAddress(updatedEmailAddress)

	if u.EmailAddress == emailAddress {
		t.Fatal("test failed attempting to call method: UpdateEmailAddress")
	}
}

func TestThatSetPasswordHashSetsPasswordHash(t *testing.T) {
	id := ulid.Make()
	firstName := "Leeroy"
	lastName := "Jenkins"
	emailAddress := "ljenkins@example.com"
	u := Register(id, firstName, lastName, emailAddress)

	p := "pa$$w0rd"
	ph := password.Create(p)
	u.SetPasswordHash(ph)

	if u.PasswordHash.Compare(p) == false {
		t.Fatal("test failed attempting to call method: SetPasswordHash")
	}
}

func TestThatAddRoleAddsRole(t *testing.T) {
	id := ulid.Make()
	firstName := "Leeroy"
	lastName := "Jenkins"
	emailAddress := "ljenkins@example.com"
	u := Register(id, firstName, lastName, emailAddress)

	roleName := "ROLE_SUPER_ADMIN"
	r := role.Define(ulid.Make(), roleName)
	u.AddRole(*r)

	if u.HasRole(roleName) == false {
		t.Fatal("test failed attempting to call method: AddRole")
	}
}

func TestThatAddRoleReturnsEarlyIfRoleAlreadyExists(t *testing.T) {
	id := ulid.Make()
	firstName := "Leeroy"
	lastName := "Jenkins"
	emailAddress := "ljenkins@example.com"
	u := Register(id, firstName, lastName, emailAddress)

	roleName := "ROLE_SUPER_ADMIN"
	r := role.Define(ulid.Make(), roleName)
	u.AddRole(*r)
	u.AddRole(*r)

	if u.HasRole(roleName) == false {
		t.Fatal("test failed attempting to call method: AddRole")
	}
}

func TestThatRemoveRoleRemovesRole(t *testing.T) {
	id := ulid.Make()
	firstName := "Leeroy"
	lastName := "Jenkins"
	emailAddress := "ljenkins@example.com"
	u := Register(id, firstName, lastName, emailAddress)

	roleName := "ROLE_SUPER_ADMIN"
	r := role.Define(ulid.Make(), roleName)
	u.AddRole(*r)
	u.RemoveRole(u.Roles[0])

	if u.HasRole(roleName) == true {
		t.Fatal("test failed attempting to call method: RemoveRole")
	}
}

func TestThatPermissionsReturnsPermissions(t *testing.T) {
	id := ulid.Make()
	firstName := "Leeroy"
	lastName := "Jenkins"
	emailAddress := "ljenkins@example.com"
	u := Register(id, firstName, lastName, emailAddress)

	roleName := "ROLE_SUPER_ADMIN"
	r := role.Define(ulid.Make(), roleName)
	permissionName := "CREATE_USERS"
	p := permission.Define(ulid.Make(), permissionName, "Allows the user to create users")
	r.AddPermission(*p)
	u.AddRole(*r)

	if u.HasPermission(permissionName) == false {
		t.Fatal("test failed attempting to call method: Permissions")
	}
}

func TestThatHasPermissionReturnsFalse(t *testing.T) {
	id := ulid.Make()
	firstName := "Leeroy"
	lastName := "Jenkins"
	emailAddress := "ljenkins@example.com"
	u := Register(id, firstName, lastName, emailAddress)

	if u.HasPermission("FOO_BAR") == true {
		t.Fatal("test failed attempting to call method: HasPermission")
	}
}

func TestThatDeleteDeletesUser(t *testing.T) {
	id := ulid.Make()
	firstName := "Leeroy"
	lastName := "Jenkins"
	emailAddress := "ljenkins@example.com"
	u := Register(id, firstName, lastName, emailAddress)

	u.Delete()

	if u.IsDeleted() == false {
		t.Fatal("test failed attempting to call method: Delete")
	}
}

func TestThatRestoreRestoresUser(t *testing.T) {
	id := ulid.Make()
	firstName := "Leeroy"
	lastName := "Jenkins"
	emailAddress := "ljenkins@example.com"
	u := Register(id, firstName, lastName, emailAddress)

	u.Delete()
	u.Restore()

	if u.IsDeleted() == true {
		t.Fatal("test failed attempting to call method: Restore")
	}
}
