package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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
	router.Route("/api", func(router chi.Router) {
		router.Post("/users/register", cfg.RegisterHandler)
		router.Post("/users/login", cfg.LoginHandler)
		router.Post("/upload", cfg.UploadFile)

		router.Route("/collections", func(r chi.Router) {
			r.Use(JWTMiddleware(cfg.JWTKey))
			r.Post("/", cfg.CreateCollection)
			r.Get("/", cfg.ListCollections)

			r.Route("/{collectionID}", func(r chi.Router) {
				r.Get("/", cfg.GetCollection)
				r.Get("/{cardID}", cfg.GetCard)
				r.Post("/", cfg.CreateCard)
			})
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
	userID, err := getIdFromContext(r.Context(), "userID")
	if err != nil {
		RespondWithErr(w, 500, err.Error())
		return
	}
	var collection server.Collection
	err = json.NewDecoder(r.Body).Decode(&collection)
	if err != nil {
		RespondWithErr(w, 500, "error on decoding")
		return
	}
	if collection.Name == "" {
		RespondWithErr(w, 400, "name cannot be empty")
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
	userID, err := getIdFromContext(r.Context(), "userID")
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
	userID, err := getIdFromContext(r.Context(), "userID")
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

func (cfg *Config) GetCard(w http.ResponseWriter, r *http.Request) {
	userID, err := getIdFromContext(r.Context(), "userID")
	if err != nil {
		RespondWithErr(w, 400, err.Error())
		return
	}

	cardId, err := getIdFromPath(r, "cardID")
	if err != nil {
		RespondWithErr(w, 400, err.Error())
		return
	}

	card, err := cfg.Server.GetCard(r.Context(), userID, cardId)
	if err != nil {
		RespondWithErr(w, 400, err.Error())
		return
	}

	RespondWithJson(w, 200, card)

}

func (cfg *Config) CreateCard(w http.ResponseWriter, r *http.Request) {

	userID, err := getIdFromContext(r.Context(), "userID")
	if err != nil {
		RespondWithErr(w, 400, err.Error())
		return
	}

	collectionID, err := getIdFromPath(r, "collectionID")
	if err != nil {
		RespondWithErr(w, 400, err.Error())
		return
	}

	//get card data from requst
	var card server.Card
	err = json.NewDecoder(r.Body).Decode(&card)
	if err != nil {
		RespondWithErr(w, 400, err.Error())
		return
	}
	if len(card.Front) == 0 || len(card.Back) == 0 {
		RespondWithErr(w, 400, "Card data cannot be empty")
		return
	}

	//check that user owns the collection
	err = cfg.Server.CheckUserOwnership(r.Context(), collectionID, userID)
	if err != nil {
		RespondWithErr(w, 403, err.Error())
		return
	}

	//create card
	err = cfg.Server.CreateCard(r.Context(), collectionID, &card)
	if err != nil {
		RespondWithErr(w, 500, err.Error())
		return
	}
	RespondWithJson(w, 200, card)
}

func getIdFromContext(ctx context.Context, key string) (uuid.UUID, error) {
	idStr := ctx.Value(key).(string)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid %v context value: %v", key, err)
	}
	return id, nil
}

func getIdFromPath(r *http.Request, key string) (uuid.UUID, error) {

	idStr := chi.URLParam(r, key)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid %v path parameter: %v", key, err)
	}
	return id, nil
}
