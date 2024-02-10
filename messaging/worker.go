package messaging

import (
	"bytes"
	"fmt"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"os"
	"time"
)

type Worker interface {
	OpenChannel() error
	DeclareQueue(queueName string) error
	Consume() error
}

type RabbitMQWorker struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
	queue      amqp091.Queue
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

	return &RabbitMQWorker{connection: conn}, nil
}

func (w *RabbitMQWorker) OpenChannel() error {
	ch, err := w.connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open server channel for message queue: %s", err)
	}

	w.channel = ch

	return nil
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

	w.queue = q

	return nil
}

func (w *RabbitMQWorker) Consume() error {
	err := w.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %s", err)
	}

	msgs, err := w.channel.Consume(
		w.queue.Name, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %s", err)
	}

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			dotCount := bytes.Count(d.Body, []byte("."))
			t := time.Duration(dotCount)
			time.Sleep(t * time.Second)
			log.Printf("Done")
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

	return nil
}

func WorkerInit() {
	w, err := NewRabbitMQWorker()
	if err != nil {
		log.Fatalf("failed to create RabbitMQ worker: %s", err)
	}
	defer w.connection.Close()

	err = w.OpenChannel()
	if err != nil {
		log.Fatalf("failed to open channel: %s", err)
	}
	defer w.channel.Close()

	err = w.DeclareQueue("task_queue")
	if err != nil {
		log.Fatalf("failed to declare queue: %s", err)
	}

	err = w.Consume()
	if err != nil {
		log.Fatalf("failed to consume messages from worker: %s", err)
	}
}
