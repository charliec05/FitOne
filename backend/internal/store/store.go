package store

import (
	"database/sql"
	"fmt"

	"fitonex/backend/internal/config"
	"fitonex/backend/internal/store/checkins"
	"fitonex/backend/internal/store/exercises"
	"fitonex/backend/internal/store/gyms"
	"fitonex/backend/internal/store/machines"
	"fitonex/backend/internal/store/moderation"
	"fitonex/backend/internal/store/migrations"
	"fitonex/backend/internal/store/users"
	"fitonex/backend/internal/store/videos"
	"fitonex/backend/internal/store/workouts"

	_ "github.com/lib/pq"
)

// Store represents the data store
type Store struct {
	db     *sql.DB
	config *config.Config
	Users      *users.Store
	Workouts   *workouts.Store
	Gyms       *gyms.Store
	Machines   *machines.Store
	Videos     *videos.Store
	Checkins   *checkins.Store
	Exercises  *exercises.Store
	Moderation *moderation.Store
}

// New creates a new store instance
func New(cfg *config.Config) *Store {
	return &Store{
		config: cfg,
	}
}

// Connect establishes a connection to the database
func (s *Store) Connect() error {
	db, err := sql.Open("postgres", s.config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	s.db = db

	// Initialize store components
	s.Users = users.New(s.db)
	s.Workouts = workouts.New(s.db)
	s.Gyms = gyms.New(s.db)
	s.Machines = machines.New(s.db)
	s.Videos = videos.New(s.db)
	s.Checkins = checkins.New(s.db)
	s.Exercises = exercises.New(s.db)
	s.Moderation = moderation.New(s.db)

	return nil
}

// Close closes the database connection
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *Store) Ping() error {
	if s.db == nil {
		return fmt.Errorf("database not connected")
	}
	return s.db.Ping()
}

// MigrateUp runs all pending migrations
func (s *Store) MigrateUp() error {
	return migrations.Up(s.db)
}

// MigrateDown rolls back the last migration
func (s *Store) MigrateDown() error {
	return migrations.Down(s.db)
}
