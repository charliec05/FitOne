package workouts

import (
	"database/sql"
	"fmt"
	"time"

	"fitonex/backend/internal/models"

	"github.com/google/uuid"
)

// Store handles workout-related database operations
type Store struct {
	db *sql.DB
}

// New creates a new workouts store
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// Create creates a new workout
func (s *Store) Create(userID, name, description string, duration int, workoutType string) (*models.Workout, error) {
	workout := &models.Workout{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        name,
		Description: description,
		Duration:    duration,
		Type:        workoutType,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO workouts (id, user_id, name, description, duration, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := s.db.Exec(query, workout.ID, workout.UserID, workout.Name, workout.Description, workout.Duration, workout.Type, workout.CreatedAt, workout.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create workout: %w", err)
	}

	return workout, nil
}

// GetByID retrieves a workout by ID
func (s *Store) GetByID(id, userID string) (*models.Workout, error) {
	workout := &models.Workout{}
	query := `
		SELECT id, user_id, name, description, duration, type, created_at, updated_at 
		FROM workouts 
		WHERE id = $1 AND user_id = $2
	`

	err := s.db.QueryRow(query, id, userID).Scan(
		&workout.ID, &workout.UserID, &workout.Name, &workout.Description, 
		&workout.Duration, &workout.Type, &workout.CreatedAt, &workout.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workout not found")
		}
		return nil, fmt.Errorf("failed to get workout: %w", err)
	}

	return workout, nil
}

// GetByUserID retrieves workouts for a user with cursor-based pagination
func (s *Store) GetByUserID(userID, cursor string, limit int) ([]models.Workout, string, error) {
	var query string
	var args []interface{}

	if cursor == "" {
		// First page
		query = `
			SELECT id, user_id, name, description, duration, type, created_at, updated_at
			FROM workouts 
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []interface{}{userID, limit + 1}
	} else {
		// Subsequent pages
		query = `
			SELECT id, user_id, name, description, duration, type, created_at, updated_at
			FROM workouts 
			WHERE user_id = $1 AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3
		`
		args = []interface{}{userID, cursor, limit + 1}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query workouts: %w", err)
	}
	defer rows.Close()

	var workouts []models.Workout
	var nextCursor string

	for rows.Next() {
		var workout models.Workout
		err := rows.Scan(
			&workout.ID, &workout.UserID, &workout.Name, &workout.Description,
			&workout.Duration, &workout.Type, &workout.CreatedAt, &workout.UpdatedAt,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan workout: %w", err)
		}

		workouts = append(workouts, workout)
	}

	// Check if there are more results
	if len(workouts) > limit {
		// Remove the extra item and set cursor for next page
		nextCursor = workouts[limit-1].CreatedAt.Format(time.RFC3339Nano)
		workouts = workouts[:limit]
	}

	return workouts, nextCursor, nil
}

// Update updates a workout
func (s *Store) Update(id, userID, name, description string, duration int, workoutType string) (*models.Workout, error) {
	query := `
		UPDATE workouts 
		SET name = $1, description = $2, duration = $3, type = $4, updated_at = $5
		WHERE id = $6 AND user_id = $7
		RETURNING id, user_id, name, description, duration, type, created_at, updated_at
	`

	workout := &models.Workout{}
	err := s.db.QueryRow(query, name, description, duration, workoutType, time.Now(), id, userID).Scan(
		&workout.ID, &workout.UserID, &workout.Name, &workout.Description,
		&workout.Duration, &workout.Type, &workout.CreatedAt, &workout.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update workout: %w", err)
	}

	return workout, nil
}

// Delete deletes a workout
func (s *Store) Delete(id, userID string) error {
	query := `DELETE FROM workouts WHERE id = $1 AND user_id = $2`
	result, err := s.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete workout: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("workout not found")
	}

	return nil
}
