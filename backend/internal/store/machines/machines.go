package machines

import (
	"database/sql"
	"fmt"
	"strings"

	"fitonex/backend/internal/models"
	"fitonex/backend/internal/pagination"
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

func (s *Store) SearchWithScores(term string, limit int, cursor *pagination.ScoreDescCursor, prefix bool) (pagination.Paginated[models.MachineSearchResult], error) {
	if limit <= 0 {
		return pagination.Paginated[models.MachineSearchResult]{}, pagination.ErrInvalidLimit
	}
	term = strings.TrimSpace(term)
	if term == "" {
		return pagination.Paginated[models.MachineSearchResult]{}, nil
	}

	var (
		query string
		args  []any
	)

	if prefix {
		query = `
SELECT id, name, body_part, 1.0 AS score
FROM machines
WHERE LOWER(name) LIKE LOWER($1) || '%'
`
		args = append(args, term)
		if cursor != nil {
			query += ` AND id > $3`
			args = append(args, limit+1, cursor.ID)
		} else {
			args = append(args, limit+1)
		}
		query += " ORDER BY score DESC, name ASC, id ASC LIMIT $2"
	} else {
		query = `
WITH ranked AS (
    SELECT id, name, body_part, similarity(name, $1) AS score
    FROM machines
)
SELECT id, name, body_part, score
FROM ranked
WHERE score > 0.1
`
		args = append(args, term, limit+1)
		if cursor != nil {
			query += ` AND (score < $3 OR (score = $3 AND id > $4))`
			args = append(args, cursor.Score, cursor.ID)
		}
		query += " ORDER BY score DESC, id ASC LIMIT $2"
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return pagination.Paginated[models.MachineSearchResult]{}, fmt.Errorf("search machines: %w", err)
	}
	defer rows.Close()

	var results []models.MachineSearchResult
	for rows.Next() {
		var item models.MachineSearchResult
		if err := rows.Scan(&item.ID, &item.Name, &item.BodyPart, &item.Score); err != nil {
			return pagination.Paginated[models.MachineSearchResult]{}, fmt.Errorf("scan machine search: %w", err)
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return pagination.Paginated[models.MachineSearchResult]{}, err
	}

	page, err := pagination.ScoreDescPage(results, limit, func(item models.MachineSearchResult) pagination.ScoreDescCursor {
		return pagination.ScoreDescCursor{Score: item.Score, ID: item.ID}
	})
	if err != nil {
		return pagination.Paginated[models.MachineSearchResult]{}, err
	}
	return page, nil
}
