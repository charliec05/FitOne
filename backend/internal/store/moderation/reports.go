package moderation

import (
	"database/sql"
	"fmt"
	"time"

	"fitonex/backend/internal/models"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(userID, objectType, objectID, reason string) (*models.ModerationReport, error) {
	report := &models.ModerationReport{
		ID:         uuid.New().String(),
		UserID:     userID,
		ObjectType: objectType,
		ObjectID:   objectID,
		Reason:     reason,
		CreatedAt:  time.Now().UTC(),
	}

	query := `
		INSERT INTO moderation_reports (id, user_id, object_type, object_id, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if _, err := s.db.Exec(query, report.ID, report.UserID, report.ObjectType, report.ObjectID, report.Reason, report.CreatedAt); err != nil {
		return nil, fmt.Errorf("create report: %w", err)
	}
	return report, nil
}
