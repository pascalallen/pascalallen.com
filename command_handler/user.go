package command_handler

import (
	"fmt"
	"github.com/pascalallen/pascalallen.com/command"
	"github.com/pascalallen/pascalallen.com/domain/user"
	"github.com/pascalallen/pascalallen.com/event"
	"github.com/pascalallen/pascalallen.com/messaging"
	"log"
)

type RegisterUserHandler struct {
	UserRepository  user.UserRepository
	EventDispatcher messaging.EventDispatcher
}

func (h RegisterUserHandler) Handle(cmd command.Command) error {
	c, ok := cmd.(*command.RegisterUser)
	if !ok {
		return fmt.Errorf("invalid command type passed to RegisterUserHandler: %v", cmd)
	}

	u := user.Register(c.Id, c.FirstName, c.LastName, c.EmailAddress)
	u.SetPasswordHash(c.PasswordHash)

	err := h.UserRepository.Add(u)
	if err != nil {
		return fmt.Errorf("user registration failed: %s", err)
	}

	evt := event.UserRegistered{
		Id:           c.Id,
		FirstName:    c.FirstName,
		LastName:     c.LastName,
		EmailAddress: c.EmailAddress,
	}
	h.EventDispatcher.Dispatch(evt)

	return nil
}

type UpdateUserHandler struct{}

func (h UpdateUserHandler) Handle(cmd command.Command) error {
	c, ok := cmd.(*command.UpdateUser)
	if !ok {
		return fmt.Errorf("invalid command type passed to UpdateUserHandler: %v", cmd)
	}

	// TODO
	log.Printf("UpdateUserHandler executed: %v", c)

	return nil
}
