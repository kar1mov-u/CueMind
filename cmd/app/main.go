package main

import (
	"CueMind/api"
	"CueMind/internal/llm"
	"CueMind/internal/server"
	"CueMind/internal/storage"
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

	cfg := api.Config{Server: server, JWTKey: jwtKey}
	log.Println("listening on 8000")
	http.ListenAndServe(":8000", cfg.CreateEndpoints())
}
