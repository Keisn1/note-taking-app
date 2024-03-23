package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt"
)

type JWTToken string

func getTokenString(r *http.Request) (string, error) {
	tokenString, _ := r.Context().Value(JWTToken("token")).(string)
	return tokenString, nil
}

func parseTokenString(tString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			slog.Info("unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		secret := []byte(os.Getenv("JWT_SECRET_KEY"))
		return secret, nil
	})
	return token, err
}

func JWTAuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, _ := getTokenString(r)
		_, err := parseTokenString(tokenString)

		if err != nil {
			http.Error(w, "Invalid Authorization", http.StatusForbidden)
			slog.Info("Invalid Authorization: %w", err)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
}
