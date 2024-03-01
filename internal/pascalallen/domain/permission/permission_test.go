package permission

import (
	"github.com/oklog/ulid/v2"
	"testing"
)

func TestThatDefineReturnsInstanceOfPermission(t *testing.T) {
	id := ulid.Make()
	name := "CREATE_USERS"
	description := "Allows the user to create users"
	p := Define(id, name, description)

	if p == nil {
		t.Fatal("test failed attempting to call method: Define")
	}
}

func TestThatUpdateNameUpdatesName(t *testing.T) {
	id := ulid.Make()
	name := "CREATE_USERS"
	description := "Allows the user to create users"
	p := Define(id, name, description)

	updatedName := "CREATE_ROLES"
	p.UpdateName(updatedName)

	if p.Name == name {
		t.Fatal("test failed attempting to call method: UpdateName")
	}
}

func TestThatUpdateDescriptionUpdatesDescription(t *testing.T) {
	id := ulid.Make()
	name := "CREATE_USERS"
	description := "Allows the user to create users"
	p := Define(id, name, description)

	updatedDescription := "Allows the user to create users."
	p.UpdateDescription(updatedDescription)

	if p.Description == description {
		t.Fatal("test failed attempting to call method: UpdateDescription")
	}
}
