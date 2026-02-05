package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"nis-pipo/internal/user"
)

type AuthHandler struct {
	service *user.Service
}

func NewAuthHandler(service *user.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if req.Email == "" || req.Password == "" {
			http.Error(w, "email and password required", http.StatusBadRequest)
			return
		}
		u, err := h.service.Register(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, user.ErrEmailExists) {
				http.Error(w, "email already exists", http.StatusConflict)
				return
			}
			http.Error(w, "cannot register", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]any{
			"id": u.ID, "email": u.Email,
		})
	})
}

func (h *AuthHandler) Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if req.Email == "" || req.Password == "" {
			http.Error(w, "email and password required", http.StatusBadRequest)
			return
		}
		u, err := h.service.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			if errors.Is(err, user.ErrInvalidCreds) {
				http.Error(w, "invalid email or password", http.StatusUnauthorized)
				return
			}
			http.Error(w, "cannot login", http.StatusInternalServerError)
			return
		}
		token, err := h.generateJWT(u.ID)
		if err != nil {
			http.Error(w, "cannot generate token", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"token": token})
	})
}

func (h *AuthHandler) generateJWT(userID string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "default-secret"
	}
	expiry := os.Getenv("JWT_EXPIRY")
	if expiry == "" {
		expiry = "24h"
	}
	d, _ := time.ParseDuration(expiry)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(d).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
