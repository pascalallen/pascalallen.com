package command_handler

import (
	"fmt"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/command"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/event"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/messaging"
)

type SendWelcomeEmailHandler struct {
	EventDispatcher messaging.EventDispatcher
}

func (h SendWelcomeEmailHandler) Handle(cmd messaging.Command) error {
	c, ok := cmd.(*command.SendWelcomeEmail)
	if !ok {
		return fmt.Errorf("invalid command type passed to SendWelcomeEmailHandler: %v", cmd)
	}

	// TODO: send welcome email

	evt := event.WelcomeEmailSent{
		Id:           c.Id,
		FirstName:    c.FirstName,
		LastName:     c.LastName,
		EmailAddress: c.EmailAddress,
		Token:        c.Token,
	}
	h.EventDispatcher.Dispatch(evt)

	return nil
}
