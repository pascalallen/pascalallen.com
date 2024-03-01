package messaging

import (
	"context"
	"encoding/json"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/command"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

type Command interface {
	CommandName() string
}

type CommandHandler interface {
	Handle(command Command) error
}

type CommandBus interface {
	RegisterHandler(commandType string, handler CommandHandler)
	StartConsuming()
	Execute(cmd Command)
}

type RabbitMqCommandBus struct {
	channel  *amqp091.Channel
	handlers map[string]CommandHandler
}

const queueName = "commands"

func NewRabbitMqCommandBus(conn *amqp091.Connection) RabbitMqCommandBus {
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

	return RabbitMqCommandBus{
		channel:  ch,
		handlers: make(map[string]CommandHandler),
	}
}

func (bus RabbitMqCommandBus) RegisterHandler(commandType string, handler CommandHandler) {
	bus.handlers[commandType] = handler
}

func (bus RabbitMqCommandBus) StartConsuming() {
	msgs := bus.messages()

	var forever chan struct{}

	go func() {
		for msg := range msgs {
			bus.processCommand(msg)
		}
	}()

	<-forever
}

func (bus RabbitMqCommandBus) Execute(cmd Command) {
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
			Type:         cmd.CommandName(),
		},
	)
	if err != nil {
		log.Fatalf("failed to publish command: %s", err)
	}
}

func (bus RabbitMqCommandBus) messages() <-chan amqp091.Delivery {
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

func (bus RabbitMqCommandBus) processCommand(msg amqp091.Delivery) {
	var cmd Command

	switch msg.Type {
	case command.RegisterUser{}.CommandName():
		cmd = &command.RegisterUser{}
	case command.UpdateUser{}.CommandName():
		cmd = &command.UpdateUser{}
	case command.SendWelcomeEmail{}.CommandName():
		cmd = &command.SendWelcomeEmail{}
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
