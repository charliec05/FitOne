package machines

import (
	"database/sql"
	"fmt"
	"strings"

	"fitonex/backend/internal/models"
)

// Store handles machine-related database operations
type Store struct {
	db *sql.DB
}

// New creates a new machines store
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// Search retrieves machines with optional filtering
func (s *Store) Search(query, bodyPart string, limit int) ([]models.Machine, error) {
	var whereClause []string
	var args []interface{}
	argIndex := 1

	// Base query
	sqlQuery := `
		SELECT id, name, body_part, created_at
		FROM machines
	`

	// Add search conditions
	if query != "" {
		whereClause = append(whereClause, fmt.Sprintf("LOWER(name) LIKE LOWER($%d)", argIndex))
		args = append(args, "%"+query+"%")
		argIndex++
	}

	if bodyPart != "" {
		whereClause = append(whereClause, fmt.Sprintf("LOWER(body_part) = LOWER($%d)", argIndex))
		args = append(args, bodyPart)
		argIndex++
	}

	// Add WHERE clause if needed
	if len(whereClause) > 0 {
		sqlQuery += " WHERE " + strings.Join(whereClause, " AND ")
	}

	// Add ordering and limit
	sqlQuery += " ORDER BY name ASC LIMIT $" + fmt.Sprintf("%d", argIndex)
	args = append(args, limit)

	rows, err := s.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search machines: %w", err)
	}
	defer rows.Close()

	var machines []models.Machine
	for rows.Next() {
		var machine models.Machine

		err := rows.Scan(
			&machine.ID, &machine.Name, &machine.BodyPart, &machine.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan machine: %w", err)
		}

		machines = append(machines, machine)
	}

	return machines, nil
}

// GetByID retrieves a machine by ID
func (s *Store) GetByID(id string) (*models.Machine, error) {
	query := `
		SELECT id, name, body_part, created_at
		FROM machines
		WHERE id = $1
	`

	var machine models.Machine
	err := s.db.QueryRow(query, id).Scan(
		&machine.ID, &machine.Name, &machine.BodyPart, &machine.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("machine not found")
		}
		return nil, fmt.Errorf("failed to get machine: %w", err)
	}

	return &machine, nil
}

// GetBodyParts retrieves all unique body parts
func (s *Store) GetBodyParts() ([]string, error) {
	query := `
		SELECT DISTINCT body_part
		FROM machines
		ORDER BY body_part
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query body parts: %w", err)
	}
	defer rows.Close()

	var bodyParts []string
	for rows.Next() {
		var bodyPart string

		err := rows.Scan(&bodyPart)
		if err != nil {
			return nil, fmt.Errorf("failed to scan body part: %w", err)
		}

		bodyParts = append(bodyParts, bodyPart)
	}

	return bodyParts, nil
}
