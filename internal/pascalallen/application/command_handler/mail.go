package command_handler

import (
	"fmt"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/command"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/infrastructure/messaging"
	"log"
)

type SendWelcomeEmailHandler struct{}

func (h SendWelcomeEmailHandler) Handle(cmd messaging.Command) error {
	c, ok := cmd.(*command.SendWelcomeEmail)
	if !ok {
		return fmt.Errorf("invalid command type passed to SendWelcomeEmailHandler: %v", cmd)
	}

	// TODO
	log.Printf("SendWelcomeEmailHandler executed: %v", c)

	return nil
}
