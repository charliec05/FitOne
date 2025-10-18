package main

import (
	"context"
	"database/sql"
	"log"

	"fitonex/backend/internal/config"
	"fitonex/backend/internal/devseed"

	_ "github.com/lib/pq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("failed to load configuration:", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("failed to ping database:", err)
	}

	if err := devseed.Seed(context.Background(), db); err != nil {
		log.Fatal("failed to seed database:", err)
	}

	log.Println("development seed completed")
}
