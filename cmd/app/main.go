package main

import (
	"log"
	"net/http"
	"github.com/joho/godotenv"
	"nis-pipo/internal/db"
	"nis-pipo/internal/repository/postgres"
	"nis-pipo/internal/transport"
	"nis-pipo/internal/user"
	"os"
)

func main() {
	godotenv.Load()

	dsn := os.Getenv("DB_DSN")
	dbx, err := db.Connect(dsn)

	if err != nil {
		log.Fatal(err)
		
	}

	err = db.Migrate(dbx)
	if err != nil {
		log.Fatal(err)
	}

	uRepo := postgres.NewUserRepo(dbx)
	uService := user.NewService(uRepo)

	handler := transport.NewUserHandler(uService)
	handler.InitHandling()
	http.ListenAndServe(":8080", nil)
}