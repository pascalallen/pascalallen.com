package command

import (
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/crypto"
	"reflect"
)

type SendWelcomeEmail struct {
	Id           ulid.ULID
	FirstName    string
	LastName     string
	EmailAddress string
	Token        crypto.Crypto
}

func (c SendWelcomeEmail) CommandName() string {
	return reflect.TypeOf(c).Name()
}
