package event

import (
	"github.com/oklog/ulid/v2"
	"reflect"
)

type UserRegistered struct {
	Id           ulid.ULID `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	EmailAddress string    `json:"email_address"`
}

func (e UserRegistered) EventName() string {
	return reflect.TypeOf(e).Name()
}

type UserUpdated struct {
	Id           ulid.ULID `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	EmailAddress string    `json:"email_address"`
}

func (e UserUpdated) EventName() string {
	return reflect.TypeOf(e).Name()
}
