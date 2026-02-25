package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"nis-pipo/internal/middleware"
	"nis-pipo/internal/user"
)

type AuthHandler struct {
	service *user.Service
}

type AuthRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthRegisterResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type AuthLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthLoginResponse struct {
	Token string `json:"token"`
}

func NewAuthHandler(service *user.Service) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register godoc
//
//	@Summary	Register
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		AuthRegisterRequest	true	"email, password"
//	@Success	201		{object}	AuthRegisterResponse
//	@Failure	400		"bad request"
//	@Failure	409		"user already exists"
//	@Router		/api/auth/register [post]
func (h *AuthHandler) Register() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req AuthRegisterRequest
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

// Login godoc
//
//	@Summary	Login
//	@Tags		auth
//	@Accept		json
//	@Produce	json
//	@Param		body	body		AuthLoginRequest	true	"email, password"
//	@Success	200		{object}	AuthLoginResponse
//	@Failure	400		"bad request"
//	@Failure	401		"invalid email or password"
//	@Router		/api/auth/login [post]
func (h *AuthHandler) Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req AuthLoginRequest
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
		token, err := h.generateJWT(u.ID, u.Email)
		if err != nil {
			http.Error(w, "cannot generate token", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"token": token})
	})
}

// Me godoc
//
//	@Summary	Get current user
//	@Tags		auth
//	@Produce	json
//	@Security	BearerAuth
//	@Success	200	{object}	map[string]string
//	@Failure	401	"unauthorized"
//	@Router		/api/auth/me [get]
func (h *AuthHandler) Me() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCors(w)
		userID, _ := r.Context().Value(middleware.UserIDKey).(string)
		if userID == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		u, err := h.service.GetByID(r.Context(), userID)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"id": u.ID, "email": u.Email})
	})
}

func (h *AuthHandler) generateJWT(userID, email string) (string, error) {
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
		"email":   email,
		"exp":     time.Now().Add(d).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
