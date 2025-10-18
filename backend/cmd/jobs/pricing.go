package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"fitonex/backend/internal/config"

	_ "github.com/lib/pq"
)

const pricingInterval = 15 * time.Minute

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	log.Println("starting pricing cache job")
	if err := recomputePriceCache(context.Background(), db); err != nil {
		log.Printf("initial price cache error: %v", err)
	}

	ticker := time.NewTicker(pricingInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if err := recomputePriceCache(ctx, db); err != nil {
			log.Printf("price cache error: %v", err)
		}
		cancel()
	}
}

func recomputePriceCache(ctx context.Context, db *sql.DB) error {
	query := `
		INSERT INTO gym_price_cache (gym_id, price_from_cents, updated_at)
		SELECT gym_id, MIN(price_cents), NOW()
		FROM gym_prices
		GROUP BY gym_id
		ON CONFLICT (gym_id) DO UPDATE
		SET price_from_cents = EXCLUDED.price_from_cents,
		    updated_at = EXCLUDED.updated_at
	`
	_, err := db.ExecContext(ctx, query)
	return err
}
