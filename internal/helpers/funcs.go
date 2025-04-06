package helpers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"

	"golang.org/x/crypto/bcrypt"
)

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
