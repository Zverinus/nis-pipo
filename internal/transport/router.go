package transport

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/swaggo/http-swagger"
	_ "nis-pipo/internal/transport/docs"
	"nis-pipo/internal/meeting"
	"nis-pipo/internal/metrics"
	"nis-pipo/internal/middleware"
	"nis-pipo/internal/participant"
	"nis-pipo/internal/user"
)

func SetupRouter(userService *user.Service, meetingService *meeting.Service, participantService *participant.Service, jwtSecret string) http.Handler {
	authHandler := NewAuthHandler(userService, jwtSecret)
	meetingHandler := NewMeetingHandler(meetingService)
	participantHandler := NewParticipantHandler(participantService)

	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Metrics)
	r.Handle("/metrics", metrics.Handler())
	r.Post("/api/auth/register", authHandler.Register().ServeHTTP)
	r.Post("/api/auth/login", authHandler.Login().ServeHTTP)
	r.With(middleware.Auth(jwtSecret)).Get("/api/auth/me", authHandler.Me().ServeHTTP)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>nis-pipo</title>
</head>
<body>
  <h1>nis-pipo</h1>
  <p>Available pages:</p>
  <ul>
    <li><a href="/swagger/">Swagger UI</a></li>
    <li><a href="/metrics">Metrics</a></li>
  </ul>
</body>
</html>`))
	})
	r.Handle("/swagger/*", httpSwagger.Handler())

	r.Route("/api/meetings", func(r chi.Router) {
		r.With(middleware.Auth(jwtSecret)).Get("/", meetingHandler.List())
		r.Get("/{id}", meetingHandler.GetByID())
		r.Post("/{id}/participants", participantHandler.Create())
		r.Put("/{id}/participants/{participant_id}/slots", participantHandler.SetSlots())
		r.With(middleware.Auth(jwtSecret)).Get("/{id}/results", meetingHandler.GetResults())
		r.With(middleware.Auth(jwtSecret)).Put("/{id}/finalize", meetingHandler.Finalize())
		r.With(middleware.Auth(jwtSecret)).Post("/", meetingHandler.Create())
		r.With(middleware.Auth(jwtSecret)).Put("/{id}", meetingHandler.Update())
		r.With(middleware.Auth(jwtSecret)).Delete("/{id}", meetingHandler.Delete())
	})

	return middleware.CORS(r)
}
