package checkins

import (
    "database/sql"
    "errors"
    "fmt"
    "time"

    "fitonex/backend/internal/models"

    "github.com/google/uuid"
)

// Store handles check-in related database operations
type Store struct {
	db *sql.DB
}

// New creates a new checkins store
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// CreateToday creates a check-in for the current UTC day. It returns the check-in and whether it was newly inserted.
func (s *Store) CreateToday(userID string) (*models.Checkin, bool, error) {
    now := time.Now().UTC()
    day := now.Truncate(24 * time.Hour)

    checkin := &models.Checkin{
        ID:     uuid.New().String(),
        UserID: userID,
        Day:    day,
    }

    query := `
        INSERT INTO checkins (id, user_id, day, created_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, day)
        DO UPDATE SET created_at = checkins.created_at
        RETURNING checkins.id, checkins.user_id, checkins.day, checkins.created_at, (xmax = 0) AS inserted
    `

    var inserted bool
    if err := s.db.QueryRow(query, checkin.ID, checkin.UserID, day, now).Scan(
        &checkin.ID,
        &checkin.UserID,
        &checkin.Day,
        &checkin.CreatedAt,
        &inserted,
    ); err != nil {
        return nil, false, fmt.Errorf("failed to create check-in: %w", err)
    }

    return checkin, inserted, nil
}

// GetStats calculates streak statistics for a user
func (s *Store) GetStats(userID string) (*models.CheckinStats, error) {
    rows, err := s.db.Query(`
        SELECT day
        FROM checkins
        WHERE user_id = $1
        ORDER BY day ASC
    `, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to query check-ins: %w", err)
    }
    defer rows.Close()

    var days []time.Time
    for rows.Next() {
        var day time.Time
        if err := rows.Scan(&day); err != nil {
            return nil, fmt.Errorf("failed to scan day: %w", err)
        }
        days = append(days, day.UTC())
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("failed to iterate check-ins: %w", err)
    }

    stats := buildCheckinStats(days)
    return &stats, nil
}

// HasCheckedInToday checks if user has checked in today
func (s *Store) HasCheckedInToday(userID string) (bool, error) {
    today := time.Now().UTC().Truncate(24 * time.Hour)

    var exists int
    err := s.db.QueryRow(`SELECT 1 FROM checkins WHERE user_id = $1 AND day = $2`, userID, today).Scan(&exists)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return false, nil
        }
        return false, fmt.Errorf("failed to check if user checked in today: %w", err)
    }

    return true, nil
}

func buildCheckinStats(days []time.Time) models.CheckinStats {
    stats := models.CheckinStats{}
    if len(days) == 0 {
        return stats
    }

    last := days[len(days)-1]
    stats.LastCheckinDay = &last

    longest := 1
    currentRun := 1
    for i := 1; i < len(days); i++ {
        if days[i].Sub(days[i-1]).Hours()/24 == 1 {
            currentRun++
        } else {
            if currentRun > longest {
                longest = currentRun
            }
            currentRun = 1
        }
    }
    if currentRun > longest {
        longest = currentRun
    }
    stats.LongestStreakDays = longest

    current := 1
    for i := len(days) - 1; i > 0; i-- {
        if days[i].Sub(days[i-1]).Hours()/24 == 1 {
            current++
        } else {
            break
        }
    }
    stats.CurrentStreakDays = current

    return stats
}

func (s *Store) ExportByUser(userID string) ([]models.Checkin, error) {
	rows, err := s.db.Query(`SELECT id, user_id, day, created_at FROM checkins WHERE user_id = $1 ORDER BY day DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.Checkin
	for rows.Next() {
		var item models.Checkin
		if err := rows.Scan(&item.ID, &item.UserID, &item.Day, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) DeleteByUser(userID string) error {
	_, err := s.db.Exec(`DELETE FROM checkins WHERE user_id = $1`, userID)
	return err
}

func (s *Store) TopStreaks(period time.Duration, limit int) ([]models.LeaderboardEntry, error) {
	if limit <= 0 {
		limit = 10
	}
	since := time.Now().UTC().Add(-period)
	rows, err := s.db.Query(`
		SELECT c.user_id, u.name, COUNT(*) AS total
		FROM checkins c
		JOIN users u ON c.user_id = u.id
		WHERE c.day >= $1 AND (u.deleted_at IS NULL)
		GROUP BY c.user_id, u.name
		ORDER BY total DESC
		LIMIT $2
	`, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.LeaderboardEntry
	for rows.Next() {
		var entry models.LeaderboardEntry
		if err := rows.Scan(&entry.UserID, &entry.UserName, &entry.StreakDays); err != nil {
			return nil, err
		}
		results = append(results, entry)
	}
	return results, rows.Err()
}
