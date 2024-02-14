package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"os"
	"reflect"
	"time"
)

type Worker interface {
	DeclareQueue(queueName string) error
	PublishMessage(queueName string, message interface{}) error
	Close()
}

type RabbitMQWorker struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
	queues     []amqp091.Queue
}

func NewRabbitMQWorker() (*RabbitMQWorker, error) {
	url := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_DEFAULT_USER"),
		os.Getenv("RABBITMQ_DEFAULT_PASS"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to message queue: %s", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open server channel for message queue: %s", err)
	}

	return &RabbitMQWorker{connection: conn, channel: ch}, nil
}

func (w *RabbitMQWorker) DeclareQueue(queueName string) error {
	q, err := w.channel.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
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
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
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
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %s", err)
	}

	return w.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
}

func (w *RabbitMQWorker) Close() {
	w.connection.Close()
	w.channel.Close()
}
