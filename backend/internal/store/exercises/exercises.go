package exercises

import (
	"database/sql"
	"fmt"
	"time"

	"fitonex/backend/internal/models"
	"fitonex/backend/internal/pagination"

	"github.com/google/uuid"
)

// Store handles exercise-related database operations
type Store struct {
	db *sql.DB
}

// New creates a new exercises store
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// Create creates a new exercise with sets.
func (s *Store) Create(userID string, performedAt time.Time, gymID, machineID *string, name string, sets []models.Set) (*models.Exercise, error) {
	exercise := &models.Exercise{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      name,
		CreatedAt: performedAt.UTC(),
	}

	if gymID != nil && *gymID != "" {
		copy := *gymID
		exercise.GymID = &copy
	}

	if machineID != nil && *machineID != "" {
		copy := *machineID
		exercise.MachineID = &copy
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

    // Insert exercise
    exerciseQuery := `
        INSERT INTO exercises (id, user_id, gym_id, machine_id, name, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `

    var gymRef interface{}
    if exercise.GymID != nil {
        gymRef = *exercise.GymID
    }

    var machineRef interface{}
    if exercise.MachineID != nil {
        machineRef = *exercise.MachineID
    }

    if _, err = tx.Exec(exerciseQuery, exercise.ID, exercise.UserID, gymRef, machineRef, exercise.Name, exercise.CreatedAt); err != nil {
        return nil, fmt.Errorf("failed to create exercise: %w", err)
    }

    // Insert sets
    for i := range sets {
        sets[i].ID = uuid.New().String()
        sets[i].ExerciseID = exercise.ID
        sets[i].SetIndex = i + 1

        set := sets[i]

        setQuery := `
            INSERT INTO sets (id, exercise_id, set_index, reps, weight_kg, rpe, notes)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
        `

        if _, err = tx.Exec(setQuery, set.ID, set.ExerciseID, set.SetIndex, set.Reps, set.WeightKg, set.RPE, set.Notes); err != nil {
            return nil, fmt.Errorf("failed to create set: %w", err)
        }
    }

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

    exercise.Sets = sets
    return exercise, nil
}

// ListByDay retrieves exercises for a specific day ordered by created_at desc.
func (s *Store) ListByDay(userID string, day time.Time, limit int, cursor *pagination.TimeDescCursor) (pagination.Paginated[models.Exercise], error) {
	if limit <= 0 {
		return pagination.Paginated[models.Exercise]{}, pagination.ErrInvalidLimit
	}

	startOfDay := day.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	query := `
		SELECT 
			e.id,
			e.user_id,
			e.gym_id,
			e.machine_id,
			e.name,
			e.created_at,
			g.name AS gym_name,
			m.name AS machine_name
		FROM exercises e
		LEFT JOIN gyms g ON e.gym_id = g.id
		LEFT JOIN machines m ON e.machine_id = m.id
		WHERE e.user_id = $1 AND e.created_at >= $2 AND e.created_at < $3
	`

	args := []interface{}{userID, startOfDay, endOfDay}

	if cursor != nil {
		query += `
			AND (
				e.created_at < $4 OR
				(e.created_at = $4 AND e.id < $5)
			)
		`
		args = append(args, cursor.CreatedAt, cursor.ID)
	}

	args = append(args, limit+1)
	limitIdx := len(args)

	query += fmt.Sprintf(`
		ORDER BY e.created_at DESC, e.id DESC
		LIMIT $%d
	`, limitIdx)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return pagination.Paginated[models.Exercise]{}, fmt.Errorf("failed to query exercises: %w", err)
	}
	defer rows.Close()

	var items []models.Exercise

	for rows.Next() {
		var (
			exercise    models.Exercise
			gymID       sql.NullString
			machineID   sql.NullString
			gymName     sql.NullString
			machineName sql.NullString
		)

		if err := rows.Scan(
			&exercise.ID,
			&exercise.UserID,
			&gymID,
			&machineID,
			&exercise.Name,
			&exercise.CreatedAt,
			&gymName,
			&machineName,
		); err != nil {
			return pagination.Paginated[models.Exercise]{}, fmt.Errorf("failed to scan exercise: %w", err)
		}

		exercise.CreatedAt = exercise.CreatedAt.UTC()

		if gymID.Valid {
			value := gymID.String
			exercise.GymID = &value
		}
		if machineID.Valid {
			value := machineID.String
			exercise.MachineID = &value
		}
		if gymName.Valid {
			value := gymName.String
			exercise.GymName = &value
		}
		if machineName.Valid {
			value := machineName.String
			exercise.MachineName = &value
		}

		sets, err := s.getSetsForExercise(exercise.ID)
		if err != nil {
			return pagination.Paginated[models.Exercise]{}, err
		}
		exercise.Sets = sets

		items = append(items, exercise)
	}

	if err := rows.Err(); err != nil {
		return pagination.Paginated[models.Exercise]{}, fmt.Errorf("exercise rows error: %w", err)
	}

	page, err := pagination.TimeDescPage(items, limit, func(item models.Exercise) pagination.TimeDescCursor {
		return pagination.TimeDescCursor{CreatedAt: item.CreatedAt, ID: item.ID}
	})
	if err != nil {
		return pagination.Paginated[models.Exercise]{}, err
	}

	return page, nil
}

// getSetsForExercise retrieves all sets for a specific exercise
func (s *Store) getSetsForExercise(exerciseID string) ([]models.Set, error) {
	query := `
		SELECT id, exercise_id, set_index, reps, weight_kg, rpe, notes
		FROM sets
		WHERE exercise_id = $1
		ORDER BY set_index
	`

	rows, err := s.db.Query(query, exerciseID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sets: %w", err)
	}
	defer rows.Close()

	var sets []models.Set
	for rows.Next() {
		var set models.Set

		err := rows.Scan(
			&set.ID, &set.ExerciseID, &set.SetIndex, &set.Reps, 
			&set.WeightKg, &set.RPE, &set.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan set: %w", err)
		}

		sets = append(sets, set)
	}

	return sets, nil
}

// GetByID retrieves an exercise by ID with its sets
func (s *Store) GetByID(id, userID string) (*models.Exercise, error) {
	query := `
		SELECT 
			e.id,
			e.user_id,
			e.gym_id,
			e.machine_id,
			e.name,
			e.created_at,
			g.name AS gym_name,
			m.name AS machine_name
		FROM exercises e
		LEFT JOIN gyms g ON e.gym_id = g.id
		LEFT JOIN machines m ON e.machine_id = m.id
		WHERE e.id = $1 AND e.user_id = $2
	`

	var (
		exercise    models.Exercise
		gymID       sql.NullString
		machineID   sql.NullString
		gymName     sql.NullString
		machineName sql.NullString
	)

	err := s.db.QueryRow(query, id, userID).Scan(
		&exercise.ID,
		&exercise.UserID,
		&gymID,
		&machineID,
		&exercise.Name,
		&exercise.CreatedAt,
		&gymName,
		&machineName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("exercise not found")
		}
		return nil, fmt.Errorf("failed to get exercise: %w", err)
	}

	exercise.CreatedAt = exercise.CreatedAt.UTC()

	if gymID.Valid {
		value := gymID.String
		exercise.GymID = &value
	}
	if machineID.Valid {
		value := machineID.String
		exercise.MachineID = &value
	}
	if gymName.Valid {
		value := gymName.String
		exercise.GymName = &value
	}
	if machineName.Valid {
		value := machineName.String
		exercise.MachineName = &value
	}

	sets, err := s.getSetsForExercise(exercise.ID)
	if err != nil {
		return nil, err
	}
	exercise.Sets = sets

	return &exercise, nil
}
