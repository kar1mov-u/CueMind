package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"

	"CueMind/internal/database"
	"CueMind/internal/server"
	queue "CueMind/internal/worker-queue"
	"CueMind/internal/ws"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"

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

				//files
				r.Get("/presigUrl", cfg.GeneratePresignedUrl)
				r.Post("/verifyUpload", cfg.VerifyUpload)
				r.Get("/files", cfg.GetFilesForCollection)

				//cards
				r.Post("/cards", cfg.CreateCard)
				r.Route("/cards/{cardID}", func(r chi.Router) {
					r.Get("/", cfg.GetCard)
					r.Delete("/", cfg.DeleteCard)
					r.Put("/", cfg.UpdateCard)
				})

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

func RespondWithJson(w http.ResponseWriter, code int, data any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log.Println("Failed on Encodeing JSON")
		http.Error(w, fmt.Sprintf("failed in returning Json err: %v", err), 500)
		return
	}
}

func RespondWithErr(w http.ResponseWriter, code int, errorString string) {
	RespondWithJson(w, code, map[string]string{"Error": errorString})
}

func HashPass(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ValidatePass(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
