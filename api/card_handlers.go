package api

import (
	"CueMind/internal/server"
	"context"
	"encoding/json"
	"net/http"
)

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

func (cfg *Config) DeleteCard(w http.ResponseWriter, r *http.Request) {
	userID, err := getIdFromContext(r.Context(), "userID")
	if err != nil {
		RespondWithErr(w, http.StatusBadRequest, err.Error())
		return
	}
	collectionID, err := getIdFromPath(r, "collectionID")
	if err != nil {
		RespondWithErr(w, http.StatusBadGateway, err.Error())
		return
	}
	cardID, err := getIdFromPath(r, "cardID")
	if err != nil {
		RespondWithErr(w, http.StatusBadRequest, err.Error())
		return
	}

	//check user owns the collection
	err = cfg.Server.CheckUserOwnership(r.Context(), collectionID, userID)
	if err != nil {
		RespondWithErr(w, 403, err.Error())
		return
	}

	err = cfg.Server.DeleteCard(context.TODO(), cardID, collectionID)
	if err != nil {
		RespondWithErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJson(w, 202, nil)
}

func (cfg *Config) UpdateCard(w http.ResponseWriter, r *http.Request) {
	userID, err := getIdFromContext(r.Context(), "userID")
	if err != nil {
		RespondWithErr(w, http.StatusBadRequest, err.Error())
		return
	}
	collectionID, err := getIdFromPath(r, "collectionID")
	if err != nil {
		RespondWithErr(w, http.StatusBadGateway, err.Error())
		return
	}
	cardID, err := getIdFromPath(r, "cardID")
	if err != nil {
		RespondWithErr(w, http.StatusBadRequest, err.Error())
		return
	}

	//check user owns the collection
	err = cfg.Server.CheckUserOwnership(r.Context(), collectionID, userID)
	if err != nil {
		RespondWithErr(w, 403, err.Error())
		return
	}

	type Data struct {
		Front string `json:"front"`
		Back  string `json:"back"`
	}
	var data Data

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		RespondWithErr(w, http.StatusBadRequest, err.Error())
		return
	}

	err = cfg.Server.UpdateCard(r.Context(), cardID, data.Front, data.Back)
	if err != nil {
		RespondWithErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	RespondWithJson(w, 204, nil)

}
