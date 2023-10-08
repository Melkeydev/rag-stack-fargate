package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	ragCrypto "ragStackFargate/clients/crypto"

	"github.com/aws/aws-lambda-go/events"
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
	var registerReq RegisterRequest

	err := json.Unmarshal([]byte(request.Body), &registerReq)
	if err != nil {
		log.Printf("Unable to unmarshal register request:%v", err)
		return events.APIGatewayProxyResponse{Body: "Invalid Request"}, err
	}

	// Validate if the user already exists in DynamoDB
	_, err = app.db.GetUser(registerReq.Username)
	if err == nil {
		return events.APIGatewayProxyResponse{Body: "Username already exists", StatusCode: http.StatusBadRequest}, nil
	}

	hashedPassword, err := ragCrypto.GeneratePassword(registerReq.Password)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "Invalid request", StatusCode: http.StatusBadRequest}, nil
	}

	accessToken, err := app.jwt.GenerateAccessToken(registerReq.Username)
	if err != nil {
		log.Print("Could not issue jwt token")
		return events.APIGatewayProxyResponse{Body: "Internal Server Error - Generating token", StatusCode: 500}, err
	}

	refreshToken, err := app.jwt.GenerateRefreshToken(registerReq.Username)
	if err != nil {
		log.Print("Could not issue refresh token")
		return events.APIGatewayProxyResponse{Body: "Internal Server Error - Generating token", StatusCode: 500}, err
	}

	// Store the username and hashed password in DynamoDB
	err = app.db.AddUserToDB(registerReq.Username, string(hashedPassword), refreshToken)
	if err != nil {
		log.Printf("Failed to add user to DynamoDB: %v", err)
		return events.APIGatewayProxyResponse{Body: "Internal Server Error - DDB", StatusCode: http.StatusInternalServerError}, nil
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

	response := events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf(`{"access_token": "%s"}`, accessToken),
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type":                     "application/json",
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Headers":     "Content-Type",
			"Access-Control-Allow-Methods":     "OPTIONS, POST, GET",
			"Access-Control-Allow-Credentials": "true",
		},
		MultiValueHeaders: map[string][]string{"Set-Cookie": {refreshCookie.String()}},
	}

	return response, nil
}
