package command_handler

import (
	"github.com/pascalallen/pascalallen.com/command"
	"log"
)

type RegisterUserHandler struct{}

func (h *RegisterUserHandler) Handle(command command.RegisterUser) {
	log.Printf("Command passed to handler: %s", command)
	// TODO
}

type UpdateUserHandler struct{}

func (h *UpdateUserHandler) Handle(command command.UpdateUser) {
	log.Printf("Command passed to handler: %s", command)
	// TODO
}
