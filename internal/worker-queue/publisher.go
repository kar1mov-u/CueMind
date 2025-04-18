package workerqueue

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const ExchangeName = "main"

type Message struct {
	UserID       uuid.UUID `json:"userID"`
	CollectionID uuid.UUID `json:"collectionID"`
	FilePath     string    `json:"filePath"`
}

type Queue struct {
	conn *amqp.Connection
}

func New(url string) (*Queue, error) {
	conn, err := createConnection(url)
	if err != nil {
		return nil, err
	}
	return &Queue{conn: conn}, nil
}

func createConnection(url string) (*amqp.Connection, error) {

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("Cannot connect to RabbitMQ server: %v", err)
	}
	return conn, nil
}

func (q *Queue) PublishTask(msg Message) error {
	ch, err := q.conn.Channel()
	if err != nil {
		return fmt.Errorf("Error creating RabbitMQ channel: %v", err)
	}

	defer ch.Close()

	//Declare exchange
	err = ch.ExchangeDeclare(
		ExchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Error creating RabbitMQ exchange: %v", err)
	}

	//Marshal
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("Error marshalling RabbitMQ payload :%v", err)
	}

	//publish message
	err = ch.Publish(
		ExchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         jsonData,
		})
	if err != nil {
		return fmt.Errorf("Error on Publishing message RabbitMQ: %v", err)
	}

	return nil

}
