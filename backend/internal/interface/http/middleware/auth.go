package middleware

import (
	"context"
	"net/http"

	"github.com/liang21/heka/internal/domain/shared"
	"github.com/liang21/heka/internal/infrastructure/auth"
	"github.com/liang21/heka/internal/interface/http/response"
)

type contextKey string

const (
	UserIDKey contextKey = "user_id"
)

type AuthMiddleware struct {
	secret string
}

func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{secret: secret}
}

func (am *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Error(w, shared.ErrAuthMissingToken)
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			response.Error(w, shared.ErrAuthInvalidToken)
			return
		}

		token := authHeader[7:]

		claims, err := auth.ParseToken(am.secret, token)
		if err != nil {
			response.Error(w, shared.ErrAuthInvalidToken)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, string(claims.UserID))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) (shared.ID, bool) {
	userID, ok := r.Context().Value(UserIDKey).(shared.ID)
	return userID, ok
}
