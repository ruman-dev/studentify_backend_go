package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"softixa-solutions.com/studentify/cmd/server"
	"softixa-solutions.com/studentify/internal/config"
	"softixa-solutions.com/studentify/internal/database"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	srv := server.New(cfg, db)

	log.Printf("listening on %s", srv.Addr())
	if err := http.ListenAndServe(srv.Addr(), srv.Router()); err != nil {
		log.Fatalf("server: %v", err)
	}
}
