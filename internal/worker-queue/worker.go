package workerqueue

import (
	"CueMind/internal/database"
	"CueMind/internal/llm"
	"CueMind/internal/storage"
	"CueMind/internal/ws"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
)

type WorkerConfig struct {
	sql     *sql.DB
	db      *database.Queries
	llm     *llm.LLMService
	storage *storage.Storage
	queue   *amqp091.Connection
	hub     *ws.WSConnHub
}

func NewWorkerConf(sql *sql.DB, db *database.Queries, llm *llm.LLMService, str *storage.Storage, queueUrl string, hub *ws.WSConnHub) *WorkerConfig {
	conn, err := amqp091.Dial(queueUrl)
	if err != nil {
		log.Fatalf("ERROR | Cannot start WorkerConf")
	}
	return &WorkerConfig{db: db, llm: llm, storage: str, queue: conn, sql: sql, hub: hub}
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
		log.Printf("Worker %d processing job", id)

		start := time.Now()

		var messageData Message
		err := json.Unmarshal(msg.Body, &messageData)
		if err != nil {
			failure(msg, fmt.Errorf("ERROR : Worker cannot Unmarshal Message DATA:%v \n", err))
			continue
		}

		//Get file from the Storage
		ctx := context.Background()
		file, err := cfg.storage.GetFile(ctx, messageData.FileKey)
		if err != nil {
			failure(msg, fmt.Errorf("Error : Worker cannot GET file from Storage : %v \n", err))
			continue
		}
		defer file.Close()

		//Send file to the LLm

		flashcards, err := cfg.llm.GenerateCardsFromFile(ctx, file)
		if err != nil {
			failure(msg, fmt.Errorf("ERROR : Worker cannot Generate LLM Response : %v \n", err))
			continue
		}

		// Save in the DB
		//iterating then inserting.
		//TO-DO : later implement batch insert
		err = insertCardsToDB(flashcards.Cards, messageData.CollectionID, cfg)
		if err != nil {
			failure(msg, fmt.Errorf("ERROR : Worker cannot Insert cards to DB: %v \n", err))
			continue
		}

		fileID, err := uuid.Parse(messageData.FileKey)
		if err != nil {
			failure(msg, fmt.Errorf("ERROR: Worker cannot Parse file ID :%v \n", err))
			continue
		}
		err = cfg.db.Processed(ctx, database.ProcessedParams{ID: fileID, Processed: true})
		if err != nil {
			failure(msg, fmt.Errorf("ERROR: Worker cannot update processed status in DB : %v \n", err))
			continue
		}

		msg.Ack(true)
		elapsed := time.Since(start)
		log.Printf("Worker %d finished job. Elapsed time: %s\n", id, elapsed)

		err = cfg.hub.Delete(fileID.String(), messageData.FileName)
		if err != nil {
			log.Printf("cannot send to the websocket : %v", err)
		}

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

func failure(msg amqp091.Delivery, err error) {

	msg.Nack(false, false)
	log.Printf("Worker failed to process message: %v", err)
}
