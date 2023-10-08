package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	ragCrypto "ragStackFargate/clients/crypto"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (app *App) TestHandler(w http.ResponseWriter, r *http.Request) {

	responseBody := map[string]string{
		"message": "Hi you have hit this route",
	}
	responseJSON, err := json.Marshal(responseBody)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, POST, GET")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)

}

func (app *App) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {

	responseBody := map[string]string{
		"message": "This is the health check for the server",
	}

	responseJSON, err := json.Marshal(responseBody)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, POST, GET")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)

}

func (app *App) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	registerRequest := &RegisterRequest{}

	err := json.NewDecoder(r.Body).Decode(registerRequest)
	if err != nil {
		log.Printf("Unable to decode register request %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Validate if the user already exists in DynamoDB
	_, err = app.db.GetUser(registerRequest.Username)
	if err != nil {
		log.Printf("User already exists in DB %v", err)
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := ragCrypto.GeneratePassword(registerRequest.Password)
	if err != nil {
		log.Printf("Error generting password %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	accessToken, err := app.jwt.GenerateAccessToken(registerRequest.Username)
	if err != nil {
		log.Printf("Could not issue jwt token %v", err)
		http.Error(w, "Could not issue new JWT token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := app.jwt.GenerateRefreshToken(registerRequest.Username)
	if err != nil {
		log.Printf("Could not issue refresh token %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store the username and hashed password in DynamoDB
	err = app.db.AddUserToDB(registerRequest.Username, string(hashedPassword), refreshToken)
	if err != nil {
		log.Printf("Failed to add user to DynamoDB: %v", err)
		http.Error(w, "Internal server error DDB", http.StatusInternalServerError)
		return
	}

	refreshCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		Path:     "/",
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, POST, GET")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	http.SetCookie(w, &refreshCookie)

	w.WriteHeader(http.StatusOK)
	responseBody := fmt.Sprintf(`{"access_token": "%s"}`, accessToken)

	w.Write([]byte(responseBody))
}
