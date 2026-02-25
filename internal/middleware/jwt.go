package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				log.Printf("[Auth] %s %s: no Authorization header", r.Method, r.URL.Path)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Printf("[Auth] %s %s: invalid Authorization format", r.Method, r.URL.Path)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			token, err := jwt.Parse(parts[1], func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				reason := "invalid token"
				if err != nil {
					if errors.Is(err, jwt.ErrTokenExpired) {
						reason = "token_expired"
					} else if errors.Is(err, jwt.ErrSignatureInvalid) {
						reason = "invalid signature (check JWT_SECRET)"
					}
					log.Printf("[Auth] %s %s: %s: %v", r.Method, r.URL.Path, reason, err)
				}
				w.Header().Set("X-Auth-Reason", reason)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				log.Printf("[Auth] %s %s: invalid claims", r.Method, r.URL.Path)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			userID, _ := claims["user_id"].(string)
			if userID == "" {
				log.Printf("[Auth] %s %s: no user_id in claims", r.Method, r.URL.Path)
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
