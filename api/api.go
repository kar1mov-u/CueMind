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
	queue "CueMind/internal/worker-queue"
	"CueMind/internal/ws"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	_ "github.com/lib/pq"
)

type Config struct {
	Queue  *queue.Queue
	Server *server.Server
	Hub    *ws.WSConnHub
	JWTKey string
}

// TO-DO  change thin in prod
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (cfg *Config) CreateEndpoints() http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Route("/api", func(router chi.Router) {
		router.Post("/users/register", cfg.RegisterHandler)
		router.Post("/users/login", cfg.LoginHandler)

		router.Route("/collections", func(r chi.Router) {
			r.Use(JWTMiddleware(cfg.JWTKey))
			r.Post("/", cfg.CreateCollection)
			r.Get("/", cfg.ListCollections)

			r.Route("/{collectionID}", func(r chi.Router) {
				r.Delete("/", cfg.DeleteCollection)
				r.Get("/", cfg.GetCollection)

				r.Get("/presigUrl", cfg.GeneratePresignedUrl)
				r.Post("/verifyUpload", cfg.VerifyUpload)

				r.Get("/{cardID}", cfg.GetCard)
				r.Post("/", cfg.CreateCard)

				r.Get("/files", cfg.GetFilesForCollection)
			})
		})

		router.Get("/ws", cfg.Sock)

	})

	return router

}

type SocketData struct {
	FileID string `json:"fileID"`
}

func (cfg *Config) Sock(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	defer c.Close()

	data := SocketData{}
	if err = c.ReadJSON(&data); err != nil {
		log.Println("cannot read socket:", err)
		return
	}
	fileID, err := uuid.Parse(data.FileID)
	if err != nil {
		log.Println("Invalid fileID: ", err)
		return
	}

	fmt.Println(fileID)

	//save to hub
	cfg.Hub.Register(fileID.String(), c)

	//keep connection alive
	for {
		_, _, err := c.NextReader()
		if err != nil {
			break
		}
	}

}

func DBConnect(dbConnString string) (*database.Queries, *sql.DB) {

	conn, err := sql.Open("postgres", dbConnString)
	if err != nil {
		log.Fatalf("Failed on connecting to DB")
	}
	err = conn.Ping()
	if err != nil {
		log.Fatalf("Failed to Pinging DB")
	}
	queries := database.New(conn)
	return queries, conn
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

func (cfg *Config) DeleteCollection(w http.ResponseWriter, r *http.Request) {
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

	err = cfg.Server.DeleteCollection(r.Context(), collectionID, userID)
	if err != nil {
		RespondWithErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	RespondWithJson(w, 204, nil)

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

// func createPath(objectKey string) string {
// 	bucket := os.Getenv("BUCKET_NAME")
// 	region := os.Getenv("AWS_REGION")

// 	s := fmt.Sprintf("https://%v.s3.%v.amazonaws.com/%v", bucket, region, objectKey)
// 	return s
// }
