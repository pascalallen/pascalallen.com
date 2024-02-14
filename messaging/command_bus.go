package messaging

import (
	"encoding/json"
	"github.com/pascalallen/pascalallen.com/command"
	"github.com/pascalallen/pascalallen.com/command_handler"
	"github.com/rabbitmq/amqp091-go"
	"log"
)

type CommandBus struct {
	worker   *RabbitMQWorker
	handlers map[string]command_handler.CommandHandler
}

func NewCommandBus(w *RabbitMQWorker) *CommandBus {
	return &CommandBus{
		worker:   w,
		handlers: make(map[string]command_handler.CommandHandler),
	}
}

// RegisterHandler registers a command handler for a specific command type
func (bus *CommandBus) RegisterHandler(commandType string, handler command_handler.CommandHandler) {
	bus.handlers[commandType] = handler
}

// StartConsuming starts consuming messages from the command queue
func (bus *CommandBus) StartConsuming() {
	err := bus.worker.DeclareQueue("commands")
	if err != nil {
		log.Fatal("Failed to declare command queue:", err)
	}

	cmdMsgs, err := bus.worker.ConsumeMessages("commands")
	if err != nil {
		log.Fatal("Failed to register command consumer:", err)
	}

	var forever chan struct{}

	go func() {
		for msg := range cmdMsgs {
			bus.processCommand(msg)
		}
	}()

	<-forever
}

func (bus *CommandBus) processCommand(msg amqp091.Delivery) {
	var cmd command.Command

	switch msg.Type {
	case "command.RegisterUser":
		cmd = &command.RegisterUser{}
	case "command.UpdateUser":
		cmd = &command.UpdateUser{}
	default:
		log.Printf("Unknown command received: %s", msg.Type)
		return
	}

	err := json.Unmarshal(msg.Body, &cmd)
	if err != nil {
		log.Println("Failed to unmarshal command:", err)
		return
	}

	handler, found := bus.handlers[cmd.GetName()]
	if !found {
		log.Printf("No handler registered for command type: %s", cmd.GetName())
		return
	}

	handler.Handle(cmd)

	msg.Ack(false)
}
