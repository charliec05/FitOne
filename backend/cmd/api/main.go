package main

import (
	"log"
	"os"

	"fitonex/backend/internal/config"
	"fitonex/backend/internal/server"
)

func main() {
	// Load configuration from environment
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Create and start the server
	srv := server.New(cfg)
	
	log.Printf("Starting server on port %s", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
