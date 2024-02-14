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

func (bus *CommandBus) RegisterHandler(commandType string, handler command_handler.CommandHandler) {
	bus.handlers[commandType] = handler
}

func (bus *CommandBus) StartConsuming() {
	err := bus.worker.DeclareQueue("commands")
	if err != nil {
		log.Fatal("Failed to declare command queue:", err)
	}

	msgs, err := bus.worker.ConsumeMessages("commands")
	if err != nil {
		log.Fatal("Failed to register command consumer:", err)
	}

	var forever chan struct{}

	go func() {
		for msg := range msgs {
			bus.processCommand(msg)
		}
	}()

	<-forever
}

func (bus *CommandBus) Execute(cmd command.Command) {
	err := bus.worker.PublishMessage("commands", cmd)
	if err != nil {
		log.Fatal(err)
	}
}

func (bus *CommandBus) processCommand(msg amqp091.Delivery) {
	var cmd command.Command

	switch msg.Type {
	case command.RegisterUser{}.CommandName():
		cmd = &command.RegisterUser{}
	case command.UpdateUser{}.CommandName():
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

	handler, found := bus.handlers[cmd.CommandName()]
	if !found {
		log.Printf("No handler registered for command type: %s", cmd.CommandName())
		return
	}

	handler.Handle(cmd)

	msg.Ack(false)
}
