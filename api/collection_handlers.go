package api

import (
	"CueMind/internal/server"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

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
