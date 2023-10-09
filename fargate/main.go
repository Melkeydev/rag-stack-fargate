package main

import (
	"fmt"
	"log"
	"net/http"
	ragDynamo "ragStackFargate/clients/dynamo"
	ragJWT "ragStackFargate/clients/jwt"
)

type App struct {
	db  ragDynamo.UserStorageDB
	jwt ragJWT.TokenValidator
}

func NewApp(db ragDynamo.UserStorageDB, jwt ragJWT.TokenValidator) *App {
	return &App{
		db:  ragDynamo.NewDynamoDBClient(),
		jwt: ragJWT.NewJWTClient(db),
	}
}

func main() {
	db := ragDynamo.NewDynamoDBClient()
	jwt := ragJWT.NewJWTClient(db)

	app := NewApp(db, jwt)

	http.HandleFunc("/", app.HealthCheckHandler)
	http.HandleFunc("/test", app.TestHandler)
	http.HandleFunc("/register", app.RegisterHandler)
	http.HandleFunc("/refresh", app.RefreshHandler)
	http.Handle("/protected", ragJWT.ValidateJWTMiddleware(http.HandlerFunc(app.ProtectedHandler)))

	port := ":8080"
	fmt.Printf("Server listening on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
