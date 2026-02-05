package transport

import (
	"net/http"

	"nis-pipo/internal/middleware"
	"nis-pipo/internal/user"
)

func SetupRouter(userService *user.Service) http.Handler {
	authHandler := NewAuthHandler(userService)

	mux := http.NewServeMux()

	mux.Handle("POST /api/auth/register", authHandler.Register())
	mux.Handle("POST /api/auth/login", authHandler.Login())

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API"))
	})

	handler := middleware.CORS(mux)
	return handler
}
