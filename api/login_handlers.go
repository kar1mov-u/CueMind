package api

import (
	"CueMind/internal/server"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/lib/pq"
)

var UniqueViolationError = pq.Error{Code: pq.ErrorCode("23505")}

func (cfg *Config) RegisterHandler(w http.ResponseWriter, r *http.Request) {

	//get the user info from body
	regData := server.RegisterData{}
	err := json.NewDecoder(r.Body).Decode(&regData)
	if err != nil {
		RespondWithErr(w, 500, err.Error())
		return
	}

	// chech for the validation of the fields
	if len(regData.Password) < 4 || !ValidateEmail(regData.Email) {
		RespondWithErr(w, 403, "Provide valid inputs")
		return
	}

	//hash the password
	hashedPass, err := HashPass(regData.Password)
	if err != nil {
		RespondWithErr(w, 500, fmt.Sprintf("error in hashing pass: %v", err))
		return
	}
	regData.Password = hashedPass

	//create a new user in DB
	user, err := cfg.Server.CraeteUser(r.Context(), regData)
	if err != nil {
		RespondWithErr(w, 403, "Email or username is already in use")
		return
	}
	RespondWithJson(w, 200, user)
}

func (cfg *Config) LoginHandler(w http.ResponseWriter, r *http.Request) {

	//get the data from body
	type LoginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	logData := LoginData{}
	err := json.NewDecoder(r.Body).Decode(&logData)
	if err != nil {
		log.Println("Failure in decoding login data: ", err)
		RespondWithErr(w, 403, err.Error())
		return
	}

	if len(logData.Email) == 0 || len(logData.Password) == 0 {
		RespondWithErr(w, 400, "login fields cannot be empty")
		return
	}

	//get entry from user
	user, err := cfg.Server.GetUser(r.Context(), logData.Email)
	if err != nil {
		RespondWithErr(w, 404, "There is no such user")
		return
	}
	if !ValidatePass(logData.Password, user.Password()) {
		RespondWithErr(w, 403, "Invalid password")
		return
	}

	// implement JWT creation
	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(cfg.JWTKey))
	if err != nil {
		RespondWithErr(w, 500, "error on signign jwt")
		return
	}

	RespondWithJson(w, 200, map[string]string{"token": signedToken})
}

//middleware for JWT

func JWTMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing token", 400)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid token format", 400)
				return
			}

			tokenStr := parts[1]

			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method")
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}

			userId, ok := claims["sub"].(string)
			if !ok {
				http.Error(w, "invalid token payload", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), "userID", userId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
