package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/turysbekovg/movie-planner/internal/service"
)

// Ключ, по которому будем сохранять ID пользователя в контексте запроса
type contextKey string

const userContextKey = contextKey("userID")

func AuthMiddleware(authSvc *service.AuthSvc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем заголовок Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			// Проверяем, что заголовок имеет формат Bearer <token>.
			headerParts := strings.Split(authHeader, " ")
			if len(headerParts) != 2 || headerParts[0] != "Bearer" {
				http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}
			tokenString := headerParts[1]

			// Проверяем токен с помощью authSvc.
			userID, err := authSvc.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Если токен валиден -> добавляем в него ID пользователя
			ctx := context.WithValue(r.Context(), userContextKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
