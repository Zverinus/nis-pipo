package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/swaggo/http-swagger"
	_ "nis-pipo/internal/transport/docs"
	"nis-pipo/internal/meeting"
	"nis-pipo/internal/middleware"
	"nis-pipo/internal/user"
)

func SetupRouter(userService *user.Service, meetingService *meeting.Service, jwtSecret string) http.Handler {
	authHandler := NewAuthHandler(userService)
	meetingHandler := NewMeetingHandler(meetingService)

	r := chi.NewRouter()
	r.Post("/api/auth/register", authHandler.Register().ServeHTTP)
	r.Post("/api/auth/login", authHandler.Login().ServeHTTP)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("API")) })
	r.Handle("/swagger/*", httpSwagger.Handler())

	r.Route("/api/meetings", func(r chi.Router) {
		r.Get("/{id}", meetingHandler.GetByID())
		r.With(middleware.Auth(jwtSecret)).Post("/", meetingHandler.Create())
		r.With(middleware.Auth(jwtSecret)).Put("/{id}", meetingHandler.Update())
		r.With(middleware.Auth(jwtSecret)).Delete("/{id}", meetingHandler.Delete())
	})

	return middleware.CORS(r)
}
