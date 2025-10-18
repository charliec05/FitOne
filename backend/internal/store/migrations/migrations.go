package migrations

import (
	"database/sql"
	"fmt"
)

// Up runs all pending migrations
func Up(db *sql.DB) error {
	// Create users table
	usersTable := `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
	`

	if _, err := db.Exec(usersTable); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create workouts table
	workoutsTable := `
		CREATE TABLE IF NOT EXISTS workouts (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			duration INTEGER NOT NULL,
			type VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
	`

	if _, err := db.Exec(workoutsTable); err != nil {
		return fmt.Errorf("failed to create workouts table: %w", err)
	}

	// Create indexes for better performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);",
		"CREATE INDEX IF NOT EXISTS idx_workouts_user_id ON workouts(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_workouts_created_at ON workouts(created_at);",
		"CREATE INDEX IF NOT EXISTS idx_workouts_user_created ON workouts(user_id, created_at);",
	}

	for _, index := range indexes {
		if _, err := db.Exec(index); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// Down rolls back the last migration
func Down(db *sql.DB) error {
	// Drop tables in reverse order
	tables := []string{
		"DROP TABLE IF EXISTS workouts;",
		"DROP TABLE IF EXISTS users;",
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
	}

	return nil
}
