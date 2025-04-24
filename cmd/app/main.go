package main

import (
	"CueMind/api"
	"CueMind/internal/llm"
	"CueMind/internal/server"
	"CueMind/internal/storage"
	workerqueue "CueMind/internal/worker-queue"
	"CueMind/internal/ws"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	llmKey := os.Getenv("LLM_KEY")
	dbUrl := os.Getenv("DB_URL")
	jwtKey := os.Getenv("JWT_KEY")
	bucketName := os.Getenv("BUCKET_NAME")
	rabbitmqURL := os.Getenv("RABBIT_MQ_URL")

	dbCon, sqlCon := api.DBConnect(dbUrl)
	storageServer := storage.New(bucketName)
	server := server.New(dbCon, storageServer)
	queue := workerqueue.New(rabbitmqURL)
	llmServer := llm.New(llmKey)
	hub := ws.New()

	//creating workers
	workerCfg := workerqueue.NewWorkerConf(sqlCon, dbCon, llmServer, storageServer, rabbitmqURL, hub)
	go func() {
		workerqueue.StartWorkers(*workerCfg, 5)
	}()

	cfg := api.Config{Server: server, JWTKey: jwtKey, Queue: queue, Hub: hub}
	log.Println("listening on 8000")
	http.ListenAndServe(":8000", cfg.CreateEndpoints())
}
