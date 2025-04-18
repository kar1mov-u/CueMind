package workerqueue

import (
	"CueMind/internal/database"
	"CueMind/internal/llm"
	"CueMind/internal/storage"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type WorkerConfig struct {
	sql     *sql.DB
	db      *database.Queries
	llm     *llm.LLMService
	storage *storage.Storage
	queue   *amqp091.Connection
}

func NewWorkerConf(sql *sql.DB, db *database.Queries, llm *llm.LLMService, str *storage.Storage, queueUrl string) (*WorkerConfig, error) {
	conn, err := amqp091.Dial(queueUrl)
	if err != nil {
		return nil, err
	}
	return &WorkerConfig{db: db, llm: llm, storage: str, queue: conn, sql: sql}, nil
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
		log.Printf("Worker %d| Msg body: %v", id, string(msg.Body))
		var messageData Message
		err := json.Unmarshal(msg.Body, &messageData)
		if err != nil {
			log.Printf("ERROR : Worker cannot Unmarshal Message DATA:%v \n", err)
			continue
		}

		//Get file from the Storage
		ctx := context.Background()
		file, err := cfg.storage.GetFile(ctx, messageData.FileKey)
		if err != nil {
			log.Printf("Error : Worker cannot GET file from Storage : %v \n", err)
			continue
		}
		defer file.Close()

		//Send file to the LLm

		flashcards, err := cfg.llm.GenerateCardsFromFile(ctx, file)
		if err != nil {
			log.Printf("ERROR : Worker cannot Generate LLM Response : %v \n", err)
			continue
		}

		// Save in the DB
		//iterating then inserting.
		//TO-DO : later implement batch insert
		err = insertCardsToDB(flashcards.Cards, messageData.CollectionID, cfg)
		if err != nil {
			log.Printf("ERROR : Worker cannot Insert cards to DB")
			continue
		}

		msg.Ack(true)

	}
	return nil

}

func insertCardsToDB(cards []llm.Card, collectionID uuid.UUID, cfg WorkerConfig) error {
	ctx := context.Background()

	tx, err := cfg.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	qtx := cfg.db.WithTx(tx)
	for i := range cards {
		front := cards[i].Front
		back := cards[i].Back
		_, err = qtx.CreateCard(ctx, database.CreateCardParams{
			Front:        front,
			Back:         back,
			CollectionID: collectionID,
		})
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()

}
