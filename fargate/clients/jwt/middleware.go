package jwt

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"
)

func ValidateJWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		tokenString := extractTokenFromHeader(r)
		if tokenString == "" {
			http.Error(w, "Missing Auth Token", http.StatusUnauthorized)
			return
		}

		mySigningKey := []byte("randomString") // TODO: move this to a secure storage

		token, err := jwt.ParseWithClaims(tokenString, &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			return mySigningKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or Expired Token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*MyCustomClaims)
		if !ok {
			http.Error(w, "Invalid or Expired Token", http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "username", claims.Username)
		r = r.WithContext(ctx)

		// Call the next handler/function
		next.ServeHTTP(w, r)
	})
}

func extractTokenFromHeader(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	var token string
	fmt.Sscanf(bearerToken, "Bearer %s", &token)
	return token
}
