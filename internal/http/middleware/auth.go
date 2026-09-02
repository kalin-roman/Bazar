package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/kalin-roman/Bazar/internal/auth"
)

type ctxKey int

const userIDKey ctxKey = 0

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorized := r.Header.Get("Authorization")
			if len(authorized) == 0 || !strings.HasPrefix(authorized, "Bearer") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(authorized, "Bearer ")

			userID, err := auth.VerifyToken(token, secret)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
