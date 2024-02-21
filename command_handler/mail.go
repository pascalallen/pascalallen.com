package command_handler

import (
	"fmt"
	"github.com/pascalallen/pascalallen.com/command"
	"log"
)

type SendWelcomeEmailHandler struct{}

func (h SendWelcomeEmailHandler) Handle(cmd command.Command) error {
	c, ok := cmd.(*command.SendWelcomeEmail)
	if !ok {
		return fmt.Errorf("invalid command type passed to SendWelcomeEmailHandler: %v", cmd)
	}

	// TODO
	log.Printf("SendWelcomeEmailHandler executed: %v", c)

	return nil
}
