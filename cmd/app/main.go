package main

import (
	"CueMind/api"
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
	bucketName := os.Getenv("BUCKET_NAME")
	rabbitmqURL := os.Getenv("RABBIT_MQ_URL")

	dbCon, _ := api.DBConnect(dbUrl)
	storage := storage.New(bucketName)
	server := server.New(dbCon, storage)
	queue := workerqueue.New(rabbitmqURL)

	cfg := api.Config{Server: server, JWTKey: jwtKey, Queue: queue}
	log.Println("listening on 8000")
	http.ListenAndServe(":8000", cfg.CreateEndpoints())
}
