package event

import (
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/crypto"
	"reflect"
)

type WelcomeEmailSent struct {
	Id           ulid.ULID
	FirstName    string
	LastName     string
	EmailAddress string
	Token        crypto.Crypto
}

func (e WelcomeEmailSent) EventName() string {
	return reflect.TypeOf(e).Name()
}
