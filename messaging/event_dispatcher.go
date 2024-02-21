package messaging

import (
	"context"
	"encoding/json"
	"github.com/pascalallen/pascalallen.com/event"
	"github.com/pascalallen/pascalallen.com/listener"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"reflect"
	"time"
)

type EventDispatcher struct {
	channel   *amqp091.Channel
	listeners map[string]listener.Listener
}

const exchangeName = "events"

func NewEventDispatcher(conn *amqp091.Connection) *EventDispatcher {
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

	return &EventDispatcher{
		channel:   ch,
		listeners: make(map[string]listener.Listener),
	}
}

func (e *EventDispatcher) RegisterListener(eventType string, listener listener.Listener) {
	e.listeners[eventType] = listener
}

func (e *EventDispatcher) StartConsuming() {
	msgs := e.messages()

	var forever chan struct{}

	go func() {
		for msg := range msgs {
			e.processEvent(msg)
		}
	}()

	<-forever
}

func (e *EventDispatcher) Dispatch(evt event.Event) {
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
			Type:        reflect.TypeOf(evt).Name(),
		},
	)
	if err != nil {
		log.Fatalf("failed to dispatch event: %s", err)
	}
}

func (e *EventDispatcher) messages() <-chan amqp091.Delivery {
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

func (e *EventDispatcher) processEvent(msg amqp091.Delivery) {
	var evt event.Event

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
