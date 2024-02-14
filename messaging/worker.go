package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"reflect"
	"time"
)

type Worker interface {
	DeclareQueue(queueName string) error
	PublishMessage(queueName string, message interface{}) error
	Stop()
}

type RabbitMQWorker struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
	queues     []amqp091.Queue
}

func NewRabbitMQWorker() *RabbitMQWorker {
	url := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_DEFAULT_USER"),
		os.Getenv("RABBITMQ_DEFAULT_PASS"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	conn, err := amqp091.Dial(url)
	if err != nil {
		log.Fatalf("failed to connect to message queue: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open server channel for message queue: %s", err)
	}

	return &RabbitMQWorker{connection: conn, channel: ch}
}

func (w *RabbitMQWorker) DeclareQueue(queueName string) error {
	q, err := w.channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return fmt.Errorf("failed to create or fetch queue: %s", err)
	}

	w.queues = append(w.queues, q)

	return nil
}

func (w *RabbitMQWorker) PublishMessage(queueName string, message interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to publish message. failed to JSON encode: %v", message)
	}

	return w.channel.PublishWithContext(ctx,
		"",
		queueName,
		false,
		false,
		amqp091.Publishing{
			DeliveryMode: amqp091.Persistent,
			ContentType:  "text/plain",
			Body:         b,
			Type:         reflect.TypeOf(message).String(),
		})
}

func (w *RabbitMQWorker) ConsumeMessages(queueName string) (<-chan amqp091.Delivery, error) {
	err := w.channel.Qos(
		1,
		0,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %s", err)
	}

	return w.channel.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}

func (w *RabbitMQWorker) Stop() {
	err := w.connection.Close()
	if err != nil {
		log.Fatalf("failed to close connection to message queue: %s", err)
	}

	err = w.channel.Close()
	if err != nil {
		log.Fatalf("failed to close message queue channel: %s", err)
	}
}
