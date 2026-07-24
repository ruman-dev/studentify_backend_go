package config

import (
	"database/sql"

	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	DB = db

	err = DB.Ping()
	if err != nil {
		log.Fatal("Database unreachable:", err)
	}

	log.Println("Connected to PostgreSQL")
}
