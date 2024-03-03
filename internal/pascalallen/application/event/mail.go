package event

import (
	"github.com/oklog/ulid/v2"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/domain/crypto"
	"reflect"
)

type WelcomeEmailSent struct {
	Id           ulid.ULID     `json:"id"`
	FirstName    string        `json:"first_name"`
	LastName     string        `json:"last_name"`
	EmailAddress string        `json:"email_address"`
	Token        crypto.Crypto `json:"token"`
}

func (e WelcomeEmailSent) EventName() string {
	return reflect.TypeOf(e).Name()
}
