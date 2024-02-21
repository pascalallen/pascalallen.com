package listener

import (
	"fmt"
	"github.com/pascalallen/pascalallen.com/command"
	"github.com/pascalallen/pascalallen.com/domain/crypto"
	"github.com/pascalallen/pascalallen.com/event"
	"github.com/pascalallen/pascalallen.com/messaging"
)

type UserRegistration struct {
	CommandBus messaging.CommandBus
}

func (l UserRegistration) Handle(evt event.Event) error {
	e, ok := evt.(*event.UserRegistered)
	if !ok {
		return fmt.Errorf("invalid event type passed to UserRegistration listener: %v", evt)
	}

	token := crypto.Generate()
	cmd := command.SendWelcomeEmail{
		Id:           e.Id,
		FirstName:    e.FirstName,
		LastName:     e.LastName,
		EmailAddress: e.EmailAddress,
		Token:        token,
	}
	l.CommandBus.Execute(cmd)

	return nil
}
