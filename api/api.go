package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"CueMind/internal/database"
	"CueMind/internal/server"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

type Config struct {
	Server *server.Server
	JWTKey string
}

func (cfg *Config) CreateEndpoints() http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	router.Route("/api", func(router chi.Router) {
		router.Post("/users/register", cfg.RegisterHandler)
		router.Post("/users/login", cfg.LoginHandler)

		router.Route("/collections", func(r chi.Router) {
			r.Use(JWTMiddleware(cfg.JWTKey))
			r.Get("/", cfg.ListCollections)
			r.Get("/{collectionID}", cfg.GetCollection)
			r.Post("/", cfg.CreateCollection)
		})

	})

	return router

	//user endpoint

	//collection endpoints

	//card endpoints
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

func (cfg *Config) CreateCollection(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithErr(w, 500, err.Error())
		return
	}
	var collection server.Collection
	err = json.NewDecoder(r.Body).Decode(&collection)
	if err != nil {
		RespondWithErr(w, 500, err.Error())
		return
	}
	err = cfg.Server.CreateCollection(r.Context(), userID, &collection)
	if err != nil {
		RespondWithErr(w, 500, err.Error())
		return
	}
	RespondWithJson(w, 200, collection)
}

func (cfg *Config) GetCollection(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithErr(w, 500, err.Error())
		return
	}

	collectionIDStr := chi.URLParam(r, "collectionID")
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		RespondWithErr(w, 500, err.Error())
		return
	}

	collection, err := cfg.Server.GetCollection(r.Context(), userID, collectionID)
	if err != nil {
		RespondWithErr(w, 500, err.Error())
		return
	}
	RespondWithJson(w, 200, collection)

}

func (cfg *Config) ListCollections(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserID(r.Context())
	if err != nil {
		RespondWithErr(w, 500, err.Error())
	}
	collections, err := cfg.Server.ListCollections(r.Context(), userID)
	if err != nil {
		RespondWithErr(w, 500, err.Error())
		return
	}
	RespondWithJson(w, 200, collections)
}

func getUserID(ctx context.Context) (uuid.UUID, error) {
	idStr := ctx.Value("userID").(string)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.UUID{}, err
	}
	return id, nil
}
