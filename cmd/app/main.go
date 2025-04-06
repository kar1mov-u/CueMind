package main

import (
	"CueMind/internal/api"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	dbUrl := os.Getenv("DB_URL")
	dbCon := api.DBConnect(dbUrl)

	cfg := api.Config{DB: dbCon}

	http.ListenAndServe(":8000", cfg.CreateEndpoints())
}
