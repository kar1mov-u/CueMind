package main

import (
	"CueMind/api"
	"CueMind/internal/llm"
	"CueMind/internal/server"
	"CueMind/internal/storage"
	workerqueue "CueMind/internal/worker-queue"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	jwtKey := os.Getenv("JWT_KEY")
	llmKey := os.Getenv("LLM_KEY")
	bucketName := os.Getenv("BUCKET_NAME")
	rabbitmqURL := os.Getenv("RABBIT_MQ_URL")
	dbCon := api.DBConnect(dbUrl)

	llmClient, err := llm.CreateClient(llmKey)
	if err != nil {
		log.Fatalf("error on creating LLMClient: %v", err)
	}

	storage, err := storage.NewStorage(bucketName)
	if err != nil {
		log.Fatalf("error on creaitng storage: %v", err)
	}

	llm := llm.NewLLMService(llmClient)
	server := server.NewServer(llm, dbCon, storage)

	queue, err := workerqueue.New(rabbitmqURL)
	if err != nil {
		log.Fatalf("error on creating qeueu: %v", err)
	}

	cfg := api.Config{Server: server, JWTKey: jwtKey, Queue: queue}
	log.Println("listening on 8000")
	http.ListenAndServe(":8000", cfg.CreateEndpoints())
}
