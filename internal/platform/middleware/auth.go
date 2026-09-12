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

// ContextWithUserID returns a copy of ctx carrying userID, the same
// way Auth's success path does. Exported alongside the getter mainly
// so other packages' handler tests can simulate "a request that
// already passed Auth" directly, without needing a real token or
// Verifier — userIDKey stays unexported either way, so nothing
// outside this package can collide with it.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func Auth(verifier *auth.Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorized := r.Header.Get("Authorization")
			if len(authorized) == 0 || !strings.HasPrefix(authorized, "Bearer") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(authorized, "Bearer ")

			userID, err := verifier.VerifyToken(token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
