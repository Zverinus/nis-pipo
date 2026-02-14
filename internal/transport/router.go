package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/swaggo/http-swagger"
	_ "nis-pipo/internal/transport/docs"
	"nis-pipo/internal/middleware"
	"nis-pipo/internal/user"
)

func SetupRouter(userService *user.Service) http.Handler {
	authHandler := NewAuthHandler(userService)

	r := chi.NewRouter()
	r.Post("/api/auth/register", authHandler.Register().ServeHTTP)
	r.Post("/api/auth/login", authHandler.Login().ServeHTTP)

	r.Handle("/swagger/*", httpSwagger.Handler())

	return middleware.CORS(r)
}
