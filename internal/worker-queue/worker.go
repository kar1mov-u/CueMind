package workerqueue

import (
	"CueMind/internal/database"
	"CueMind/internal/llm"
	"CueMind/internal/storage"
	"fmt"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type WorkerConfig struct {
	db      *database.Queries
	llm     *llm.LLMService
	storage *storage.Storage
	queue   *amqp091.Connection
}

func NewWorkerConf(db *database.Queries, llm *llm.LLMService, str *storage.Storage, queueUrl string) (*WorkerConfig, error) {
	conn, err := amqp091.Dial(queueUrl)
	if err != nil {
		return nil, err
	}
	return &WorkerConfig{db: db, llm: llm, storage: str, queue: conn}, nil
}

func StartWorkers(cfg WorkerConfig, n int) {
	for i := 1; i < n; i++ {
		go func(id int) {
			if err := startSingleWorker(id, cfg); err != nil {
				log.Printf("Worker %d crashed :%v", id, err)
			}
		}(i)

	}
}
func startSingleWorker(id int, cfg WorkerConfig) error {
	ch, err := cfg.queue.Channel()
	if err != nil {
		return fmt.Errorf("Cannot create channel from RabbitMQ: %v", err)
	}
	defer ch.Close()

	//craete Queue
	q, err := ch.QueueDeclare(
		"file-processing",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Cannot declare a queue:%v", err)
	}

	//Bind Queue to exchange
	err = ch.QueueBind(
		q.Name,
		"",
		"main",
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Cannot Bind Queue to exchange:%v", err)
	}

	ch.Qos(1, 0, false)
	//Consume

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("Cannot Consume Messages: %v", err)
	}

	for msg := range msgs {
		log.Printf("Worker %d| Msg body: %v", id, msg.Body)
	}
	return nil

}
