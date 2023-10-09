package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	ragCrypto "ragStackFargate/clients/crypto"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
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
	if err == nil {
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

func (app *App) LoginHandler(w http.ResponseWriter, r *http.Request) {
	loginRequest := &LoginRequest{}

	err := json.NewDecoder(r.Body).Decode(loginRequest)
	if err != nil {
		log.Printf("Unable to decode login request:%v", err)
		http.Error(w, "Unable to decode login request", http.StatusInternalServerError)
		return
	}

	// Check if the user exists in DynamoDB
	user, err := app.db.GetUser(loginRequest.Username)
	if err != nil {
		log.Printf("Failed to retrieve user from DynamoDB: %v", err)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Check password is correct for user
	if !ragCrypto.ComparePasswords(user.Password, loginRequest.Password) {
		log.Printf("Passwords did not match %v", err)
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Create new access token
	accessToken, err := app.jwt.GenerateAccessToken(loginRequest.Username)
	if err != nil {
		log.Print("Could not issue jwt token")
		http.Error(w, "Could not issue JWT token", http.StatusInternalServerError)
		return
	}

	// create new refresh token
	refreshToken, err := app.jwt.GenerateRefreshToken(loginRequest.Username)
	if err != nil {
		log.Print("Could not issue refresh token")
		http.Error(w, "Could not issue refresh token", http.StatusInternalServerError)
		return
	}

	// update refresh token in DynamoDB
	err = app.db.UpdateUserToken(loginRequest.Username, refreshToken)
	if err != nil {
		log.Printf("Failed to update user's token in DynamoDB: %v", err)
		http.Error(w, "Failed to update user token in database", http.StatusInternalServerError)
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

func (app *App) ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("Authorization")

	if username == "" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	responseBody := fmt.Sprintf("Hey %s - this is a protected route", username)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(responseBody))
}

func extractRefreshTokenFromCookie(r *http.Request) (string, error) {
	cookieHeader := r.Header.Get("Cookie")
	cookies := strings.Split(cookieHeader, "; ")
	for _, cookie := range cookies {
		parts := strings.SplitN(cookie, "=", 2)
		if len(parts) == 2 && parts[0] == "refresh_token" {
			return parts[1], nil
		}
	}
	return "", errors.New("Refresh Token not found in cookies")
}

func (app *App) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := extractRefreshTokenFromCookie(r)
	if err != nil {
		log.Printf("Problem extracting refresh token %v", err)
		http.Error(w, "Problem extracting refresh token", http.StatusInternalServerError)
		return
	}

	username, err := app.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		log.Printf("Invalid refresh token %v", err)
		http.Error(w, "Invalid refresh token", http.StatusInternalServerError)
		return
	}

	accessToken, err := app.jwt.GenerateAccessToken(username)
	if err != nil {
		log.Printf("Internal server error %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create an HTTP-only cookie for the new refresh token
	newRefreshToken, err := app.jwt.GenerateRefreshToken(username)
	if err != nil {
		log.Printf("Problem generating refresh token %v", err)
		http.Error(w, "Problem generating refresh token", http.StatusInternalServerError)
		return
	}

	refreshCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
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

	responseBody := fmt.Sprintf(`{"access_token":"%s"}`, accessToken)
	w.Write([]byte(responseBody))

}
