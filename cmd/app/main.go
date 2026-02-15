package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"nis-pipo/internal/db"
	"nis-pipo/internal/meeting"
	"nis-pipo/internal/repository/postgres"
	"nis-pipo/internal/transport"
	"nis-pipo/internal/user"
)

func main() {
	godotenv.Load()

	dsn := os.Getenv("DB_DSN")
	dbx, err := db.Connect(dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Migrate(dbx); err != nil {
		log.Fatal(err)
	}

	uRepo := postgres.NewUserRepo(dbx)
	uService := user.NewService(uRepo)
	meetingRepo := postgres.NewMeetingRepo(dbx)
	meetingService := meeting.NewService(meetingRepo)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret"
	}

	handler := transport.SetupRouter(uService, meetingService, jwtSecret)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on :%s", port)
	log.Printf("Swagger: http://localhost:%s/swagger/index.html", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}