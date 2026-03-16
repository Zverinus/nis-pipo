package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"nis-pipo/internal/db"
	"nis-pipo/internal/meeting"
	"nis-pipo/internal/participant"
	"nis-pipo/internal/repository/postgres"
	"nis-pipo/internal/transport"
	"nis-pipo/internal/user"
)

func main() {
	godotenv.Load()

	dsn := os.Getenv("DB_DSN")
	var (
		dbx *sql.DB
		err error
	)

	for i := 0; i < 10; i++ {
		dbx, err = db.Connect(dsn)
		if err == nil {
			break
		}
		log.Printf("db connect attempt %d/10 failed: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Migrate(dbx); err != nil {
		log.Fatal(err)
	}

	uRepo := postgres.NewUserRepo(dbx)
	uService := user.NewService(uRepo)
	meetingRepo := postgres.NewMeetingRepo(dbx)
	participantSlotsRepo := postgres.NewParticipantSlotsRepo(dbx)
	meetingService := meeting.NewService(meetingRepo, participantSlotsRepo)
	participantRepo := postgres.NewParticipantRepo(dbx)
	participantService := participant.NewService(participantRepo, meetingRepo, participantSlotsRepo)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret"
	}

	handler := transport.SetupRouter(uService, meetingService, participantService, jwtSecret)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on :%s", port)
	log.Printf("Home: http://localhost:%s/", port)
	log.Printf("Swagger: http://localhost:%s/swagger/", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
