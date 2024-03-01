package event

import (
	"github.com/oklog/ulid/v2"
	"reflect"
)

type UserRegistered struct {
	Id           ulid.ULID
	FirstName    string
	LastName     string
	EmailAddress string
}

func (e UserRegistered) EventName() string {
	return reflect.TypeOf(e).Name()
}

type UserUpdated struct {
	Id           ulid.ULID
	FirstName    string
	LastName     string
	EmailAddress string
}

func (e UserUpdated) EventName() string {
	return reflect.TypeOf(e).Name()
}
