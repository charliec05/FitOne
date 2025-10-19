package social

import (
	"database/sql"
	"fmt"
	"time"

	"fitonex/backend/internal/models"
	"fitonex/backend/internal/pagination"

	"github.com/google/uuid"
)

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateComment(videoID, userID, text string) (*models.VideoComment, error) {
	comment := &models.VideoComment{
		ID:        uuid.New().String(),
		VideoID:   videoID,
		UserID:    userID,
		Text:      text,
		CreatedAt: time.Now().UTC(),
	}
	if _, err := s.db.Exec(`
		INSERT INTO video_comments (id, video_id, user_id, comment, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, comment.ID, comment.VideoID, comment.UserID, comment.Text, comment.CreatedAt); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}
	return comment, nil
}

func (s *Store) ListByVideo(videoID string, limit int, cursor *pagination.TimeDescCursor) (pagination.Paginated[models.VideoComment], error) {
	if limit <= 0 {
		return pagination.Paginated[models.VideoComment]{}, pagination.ErrInvalidLimit
	}

	query := `
		SELECT id, video_id, user_id, comment, created_at
		FROM video_comments
		WHERE video_id = $1
	`
	args := []any{videoID}
	if cursor != nil {
		query += ` AND (created_at < $2 OR (created_at = $2 AND id < $3)) ORDER BY created_at DESC, id DESC LIMIT $4`
		args = append(args, cursor.CreatedAt, cursor.ID, limit+1)
	} else {
		query += " ORDER BY created_at DESC, id DESC LIMIT $2"
		args = append(args, limit+1)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return pagination.Paginated[models.VideoComment]{}, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	var items []models.VideoComment
	for rows.Next() {
		var item models.VideoComment
		if err := rows.Scan(&item.ID, &item.VideoID, &item.UserID, &item.Text, &item.CreatedAt); err != nil {
			return pagination.Paginated[models.VideoComment]{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return pagination.Paginated[models.VideoComment]{}, err
	}

	return pagination.TimeDescPage(items, limit, func(item models.VideoComment) pagination.TimeDescCursor {
		return pagination.TimeDescCursor{CreatedAt: item.CreatedAt, ID: item.ID}
	})
}

func (s *Store) DeleteByUser(userID string) error {
	_, err := s.db.Exec(`DELETE FROM video_comments WHERE user_id = $1`, userID)
	return err
}
