package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/swaggo/http-swagger"
	_ "nis-pipo/internal/transport/docs"
	"nis-pipo/internal/meeting"
	"nis-pipo/internal/middleware"
	"nis-pipo/internal/participant"
	"nis-pipo/internal/user"
)

func SetupRouter(userService *user.Service, meetingService *meeting.Service, participantService *participant.Service, jwtSecret string) http.Handler {
	authHandler := NewAuthHandler(userService)
	meetingHandler := NewMeetingHandler(meetingService)
	participantHandler := NewParticipantHandler(participantService)

	r := chi.NewRouter()
	r.Post("/api/auth/register", authHandler.Register().ServeHTTP)
	r.Post("/api/auth/login", authHandler.Login().ServeHTTP)
	r.With(middleware.Auth(jwtSecret)).Get("/api/auth/me", authHandler.Me().ServeHTTP)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("API")) })
	r.Handle("/swagger/*", httpSwagger.Handler())

	r.Route("/api/meetings", func(r chi.Router) {
		r.With(middleware.Auth(jwtSecret)).Get("/", meetingHandler.List())
		r.Get("/{id}", meetingHandler.GetByID())
		r.Post("/{id}/participants", participantHandler.Create())
		r.Put("/{id}/participants/{token}/slots", participantHandler.SetSlots())
		r.With(middleware.Auth(jwtSecret)).Get("/{id}/results", meetingHandler.GetResults())
		r.With(middleware.Auth(jwtSecret)).Put("/{id}/finalize", meetingHandler.Finalize())
		r.With(middleware.Auth(jwtSecret)).Post("/", meetingHandler.Create())
		r.With(middleware.Auth(jwtSecret)).Put("/{id}", meetingHandler.Update())
		r.With(middleware.Auth(jwtSecret)).Delete("/{id}", meetingHandler.Delete())
	})

	return middleware.CORS(r)
}
