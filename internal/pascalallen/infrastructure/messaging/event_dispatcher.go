package messaging

import (
	"context"
	"encoding/json"
	"github.com/pascalallen/pascalallen.com/internal/pascalallen/application/event"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

type Event interface {
	EventName() string
}

type Listener interface {
	Handle(event Event) error
}

type EventDispatcher interface {
	RegisterListener(eventType string, listener Listener)
	StartConsuming()
	Dispatch(evt Event)
}

type RabbitMqEventDispatcher struct {
	channel   *amqp091.Channel
	listeners map[string]Listener
}

const exchangeName = "events"

func NewRabbitMqEventDispatcher(conn *amqp091.Connection) RabbitMqEventDispatcher {
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open server channel for event dispatcher: %s", err)
	}

	err = ch.ExchangeDeclare(
		exchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to declare exchange: %s", err)
	}

	return RabbitMqEventDispatcher{
		channel:   ch,
		listeners: make(map[string]Listener),
	}
}

func (e RabbitMqEventDispatcher) RegisterListener(eventType string, listener Listener) {
	e.listeners[eventType] = listener
}

func (e RabbitMqEventDispatcher) StartConsuming() {
	msgs := e.messages()

	var forever chan struct{}

	go func() {
		for msg := range msgs {
			e.processEvent(msg)
		}
	}()

	<-forever
}

func (e RabbitMqEventDispatcher) Dispatch(evt Event) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b, err := json.Marshal(evt)
	if err != nil {
		log.Fatalf("failed to JSON encode event: %s", err)
	}

	err = e.channel.PublishWithContext(
		ctx,
		exchangeName,
		"",
		false,
		false,
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        b,
			Type:        evt.EventName(),
		},
	)
	if err != nil {
		log.Fatalf("failed to dispatch event: %s", err)
	}
}

func (e RabbitMqEventDispatcher) messages() <-chan amqp091.Delivery {
	err := e.channel.ExchangeDeclare(
		exchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to declare exchange: %s", err)
	}

	q, err := e.channel.QueueDeclare(
		"",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to declare queue: %s", err)
	}

	err = e.channel.QueueBind(
		q.Name,
		"",
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to bind to queue: %s", err)
	}

	d, err := e.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to consume event messages: %s", err)
	}

	return d
}

func (e RabbitMqEventDispatcher) processEvent(msg amqp091.Delivery) {
	var evt Event

	switch msg.Type {
	case event.UserRegistered{}.EventName():
		evt = &event.UserRegistered{}
	case event.UserUpdated{}.EventName():
		evt = &event.UserUpdated{}
	default:
		log.Printf("Unknown event received: %s", msg.Type)
		return
	}

	err := json.Unmarshal(msg.Body, &evt)
	if err != nil {
		log.Println("Failed to unmarshal event:", err)
		return
	}

	listener, found := e.listeners[evt.EventName()]
	if !found {
		log.Printf("No listener registered for event type: %s", evt.EventName())
		return
	}

	err = listener.Handle(evt)
	if err != nil {
		log.Printf("Error calling listener: %s", err)
		return
	}

	err = msg.Ack(false)
	if err != nil {
		log.Printf("Error acknowledging event message: %s", err)
		return
	}
}
