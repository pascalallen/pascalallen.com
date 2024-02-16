package messaging

import (
	"context"
	"encoding/json"
	"github.com/pascalallen/pascalallen.com/command"
	"github.com/pascalallen/pascalallen.com/command_handler"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"reflect"
	"time"
)

type CommandBus struct {
	channel  *amqp091.Channel
	handlers map[string]command_handler.CommandHandler
}

const queueName = "commands"

func NewCommandBus(conn *amqp091.Connection) *CommandBus {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open server channel for command queue: %s", err)
	}

	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to create or fetch queue: %s", err)
	}

	return &CommandBus{
		channel:  ch,
		handlers: make(map[string]command_handler.CommandHandler),
	}
}

func (bus *CommandBus) RegisterHandler(commandType string, handler command_handler.CommandHandler) {
	bus.handlers[commandType] = handler
}

func (bus *CommandBus) StartConsuming() {
	msgs := bus.messages()

	var forever chan struct{}

	go func() {
		for msg := range msgs {
			bus.processCommand(msg)
		}
	}()

	<-forever
}

func (bus *CommandBus) Execute(cmd command.Command) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b, err := json.Marshal(cmd)
	if err != nil {
		log.Fatalf("failed to JSON encode command: %s", err)
	}

	err = bus.channel.PublishWithContext(
		ctx,
		"",
		queueName,
		false,
		false,
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent,
			ContentType:  "text/plain",
			Body:         b,
			Type:         reflect.TypeOf(cmd).Name(),
		},
	)
	if err != nil {
		log.Fatalf("failed to publish command: %s", err)
	}
}

func (bus *CommandBus) messages() <-chan amqp091.Delivery {
	err := bus.channel.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		log.Fatalf("failed to set QoS: %s", err)
	}

	d, err := bus.channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to consume command messages: %s", err)
	}

	return d
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

	err = handler.Handle(cmd)
	if err != nil {
		log.Printf("Error calling command handler: %s", err)
		return
	}

	err = msg.Ack(false)
	if err != nil {
		log.Printf("Error acknowledging command message: %s", err)
		return
	}
}
