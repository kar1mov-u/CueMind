package main

import (
	"CueMind/api"
	"CueMind/internal/llm"
	"CueMind/internal/storage"
	workerqueue "CueMind/internal/worker-queue"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	llmKey := os.Getenv("LLM_KEY")
	bucketName := os.Getenv("BUCKET_NAME")
	rabbitmqURL := os.Getenv("RABBIT_MQ_URL")

	dbCon, sqlCon := api.DBConnect(dbUrl)
	llmServer := llm.New(llmKey)
	storageServer := storage.New(bucketName)

	workerCfg := workerqueue.NewWorkerConf(sqlCon, dbCon, llmServer, storageServer, rabbitmqURL)
	workerqueue.StartWorkers(*workerCfg, 5)
	forever := make(chan struct{})
	<-forever
}
