package api

import (
	"CueMind/internal/database"
	"CueMind/internal/helpers"
	"log"

	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lib/pq"
)

var UniqueViolationError = pq.Error{Code: pq.ErrorCode("23505")}

func (cfg *Config) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	type RegisterData struct {
		Email    string `json:"email"`
		UserName string `json:"username"`
		Password string `json:"password"`
	}

	//get the user info from body
	regData := RegisterData{}
	err := json.NewDecoder(r.Body).Decode(&regData)
	if err != nil {
		helpers.RespondWithErr(w, 500, err.Error())
		return
	}

	// chech for the validation of the fields
	if len(regData.Password) < 4 || !helpers.ValidateEmail(regData.Email) {
		helpers.RespondWithErr(w, 403, "Provide valid inputs")
		return
	}

	//hash the password
	hashedPass, err := helpers.HashPass(regData.Password)
	if err != nil {
		helpers.RespondWithErr(w, 500, fmt.Sprintf("error in hashing pass: %v", err))
		return
	}
	regData.Password = hashedPass

	//create a new user in DB
	dbUser, err := cfg.DB.CreateUser(r.Context(), database.CreateUserParams{Username: regData.UserName, Email: regData.Email, Password: regData.Password})
	if err != nil {
		helpers.RespondWithErr(w, 403, "Email or username is already in use")
		return
	}
	helpers.RespondWithJson(w, 200, dbUser)

}

func (cfg *Config) LoginHandler(w http.ResponseWriter, r *http.Request) {
	type LoginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	//get the data from body
	logData := LoginData{}
	err := json.NewDecoder(r.Body).Decode(&logData)
	if err != nil {
		log.Println("Failure in decoding login data: ", err)
		helpers.RespondWithErr(w, 403, err.Error())
		return
	}

	//get entry from user
	dbUser, err := cfg.DB.GetUser(r.Context(), logData.Email)
	if err != nil {
		helpers.RespondWithErr(w, 404, "There is no such user")
		return
	}
	if !helpers.ValidatePass(logData.Password, dbUser.Password) {
		helpers.RespondWithErr(w, 403, "Invalid password")
		return
	}
	//return bool
	helpers.RespondWithJson(w, 200, "logged in")

}
