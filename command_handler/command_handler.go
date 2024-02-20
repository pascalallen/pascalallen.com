package command_handler

import "github.com/pascalallen/pascalallen.com/command"

type CommandHandler interface {
	Handle(command command.Command) error
}
