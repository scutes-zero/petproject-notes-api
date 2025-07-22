package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"notes-api/internal/utils"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func JWTAuthMiddleware(secret []byte) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			encoder := json.NewEncoder(w)

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.WriteHeader(http.StatusUnauthorized)
				encoder.Encode(map[string]string{"Unauthorized": "Authorization header is required"})

				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				w.WriteHeader(http.StatusUnauthorized)
				encoder.Encode(map[string]string{"Unauthorized": "Invalid authorization format"})

				return
			}

			tokenString := parts[1]
			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				w.WriteHeader(http.StatusUnauthorized)
				encoder.Encode(map[string]string{"Unauthorized": "Invalid token"})

				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				encoder.Encode(map[string]string{"Unauthorized": "Invalid token claims"})

				return
			}

			userID, ok := claims["sub"].(string)
			if !ok {
				w.WriteHeader(http.StatusUnauthorized)
				encoder.Encode(map[string]string{"Unauthorized": "Invalid token subject"})

				return
			}

			ctx := context.WithValue(r.Context(), utils.UserIDKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
