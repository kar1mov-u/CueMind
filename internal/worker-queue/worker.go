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
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
			failure(msg, &cfg, messageData.FileKey, messageData.FileName, err)

			continue
		}

		//Get file from the Storage
		ctx := context.Background()
		file, err := cfg.storage.GetFile(ctx, messageData.FileKey)
		if err != nil {
			failure(msg, &cfg, messageData.FileKey, messageData.FileName, err)

			continue
		}
		defer file.Close()

		//convert to PDF if its not
		if messageData.Format != "pdf" {
			//create non-pdf file in tmp
			curDir, err := os.Getwd()
			if err != nil {
				failure(msg, &cfg, messageData.FileKey, messageData.FileName, err)

				continue
			}
			nonPdfFilename := fmt.Sprintf("%v/tmp/non-pdf/%v", curDir, messageData.FileKey)
			tmpPrevFile, err := os.Create(nonPdfFilename)
			if err != nil {
				failure(msg, &cfg, messageData.FileKey, messageData.FileName, err)

				continue
			}
			//copy data from storage file to the new file
			_, err = io.Copy(tmpPrevFile, file)
			if err != nil {
				failure(msg, &cfg, messageData.FileKey, messageData.FileName, err)

				continue
			}

			tmpPrevFile.Close()

			pdfFile, err := convertToPdf(nonPdfFilename, curDir+"/tmp/pdf/")
			if err != nil {
				failure(msg, &cfg, messageData.FileKey, messageData.FileName, err)

				continue
			}

			os.Remove(nonPdfFilename)

			defer pdfFile.Close()

			file = pdfFile

		}

		//Send file to the LLm

		flashcards, err := cfg.llm.GenerateCardsFromFile(ctx, file)
		if err != nil {
			failure(msg, &cfg, messageData.FileKey, messageData.FileName, fmt.Errorf("ERROR: Worker cannot Parse file ID :%v \n", err))

			continue
		}

		// Save in the DB
		//iterating then inserting.
		//TO-DO : later implement batch insert
		err = insertCardsToDB(flashcards.Cards, messageData.CollectionID, cfg)
		if err != nil {
			failure(msg, &cfg, messageData.FileKey, messageData.FileName, fmt.Errorf("ERROR: Worker cannot Parse file ID :%v \n", err))

			continue
		}

		fileID, err := uuid.Parse(messageData.FileKey)
		if err != nil {
			failure(msg, &cfg, messageData.FileKey, messageData.FileName, fmt.Errorf("ERROR: Worker cannot Parse file ID :%v \n", err))

			continue
		}
		err = cfg.db.Processed(ctx, database.ProcessedParams{ID: fileID, Processed: true})
		if err != nil {
			failure(msg, &cfg, messageData.FileKey, messageData.FileName, fmt.Errorf("ERROR: Worker cannot Parse file ID :%v \n", err))

			continue
		}

		msg.Ack(true)
		elapsed := time.Since(start)
		log.Printf("Worker %d finished job. Elapsed time: %s\n", id, elapsed)

		err = cfg.hub.Delete(fileID.String(), messageData.FileName, true)
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

func failure(msg amqp091.Delivery, cfg *WorkerConfig, fileID, fileName string, err error) {
	cfg.hub.Delete(fileID, fileName, false)
	msg.Nack(false, false)
	log.Printf("Worker failed to process message: %v", err)
}

func convertToPdf(filename string, outputDir string) (*os.File, error) {
	start := time.Now()
	cmd := exec.Command("libreoffice", "--headless", "--convert-to", "pdf", "--outdir", outputDir, filename)
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	baseName := filepath.Base(filename)
	generatedPdf := filepath.Join(outputDir, baseName+".pdf")

	fmt.Println(generatedPdf)

	_, err = os.Stat(generatedPdf)
	if err != nil {
		return nil, fmt.Errorf("file have not been created: %v", err)
	}

	file, err := os.Open(generatedPdf)
	if err != nil {
		return nil, fmt.Errorf("error on opening file: %v", err)
	}
	err = os.Remove(generatedPdf)
	if err != nil {
		return nil, err
	}
	elapsed := time.Since(start)
	fmt.Println("taken time:", elapsed)
	return file, nil
}
