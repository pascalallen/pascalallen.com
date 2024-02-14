package command_handler

import (
	"github.com/pascalallen/pascalallen.com/command"
	"log"
)

type CommandHandler interface {
	Handle(command command.Command)
}

type RegisterUserHandler struct{}

func (h *RegisterUserHandler) Handle(cmd command.Command) {
	registerUserCommand, ok := cmd.(*command.RegisterUser)
	if !ok {
		log.Println("Invalid command type for RegisterUserCommandHandler")
		return
	}
	log.Printf("Command passed to handler: %s", registerUserCommand)
	// TODO
}

type UpdateUserHandler struct{}

func (h *UpdateUserHandler) Handle(cmd command.Command) {
	updateUserCommand, ok := cmd.(*command.UpdateUser)
	if !ok {
		log.Println("Invalid command type for UpdateUserCommandHandler")
		return
	}
	log.Printf("Command passed to handler: %s", updateUserCommand)
	// TODO
}
