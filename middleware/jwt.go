package jwt

import (
	"github.com/Keisn1/note-taking-app/auth"
	"github.com/go-chi/chi"
	"log/slog"
	"net/http"
)

type MidHandler func(http.Handler) http.Handler

func NewJwtMidHandler(a auth.AuthInterface) MidHandler {
	m := func(next http.Handler) http.Handler {
		h := func(w http.ResponseWriter, r *http.Request) {
			userID := chi.URLParam(r, "userID")
			bearerToken := r.Header.Get("Authorization")

			_, err := a.Authenticate(userID, bearerToken)
			if err != nil {
				http.Error(w, "Failed Authentication", http.StatusForbidden)
				slog.Info("Failed Authentication: ", err)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(h)
	}
	return m
}
