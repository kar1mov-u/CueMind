package api

import (
	"database/sql"
	"log"
	"net/http"

	"CueMind/internal/database"
	// "CueMind/internal/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	_ "github.com/lib/pq"
)

type Config struct {
	DB *database.Queries
}

func (cfg *Config) CreateEndpoints() http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	//-------------------Endpoints to auth--------------------------------------------------------

	//register (to-do)
	router.Post("/users/register", cfg.RegisterHandler)
	//log-in
	router.Post("/users/login", cfg.LoginHandler)

	return router
}

func DBConnect(dbConnString string) *database.Queries {

	conn, err := sql.Open("postgres", dbConnString)
	if err != nil {
		log.Fatalf("Failed on connecting to DB")
	}
	err = conn.Ping()
	if err != nil {
		log.Fatalf("Failed to Pinging DB")
	}
	queries := database.New(conn)
	return queries
}
