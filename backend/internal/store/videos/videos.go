package videos

import (
	"database/sql"
	"errors"
	"fmt"

	"fitonex/backend/internal/models"
	"fitonex/backend/internal/pagination"

	"github.com/google/uuid"
)

// ErrVideoNotFound indicates the requested video was not found.
var ErrVideoNotFound = errors.New("video not found")

// Store handles video-related database operations.
type Store struct {
	db *sql.DB
}

// New creates a new videos store.
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// Create inserts a new instruction video record.
func (s *Store) Create(machineID, uploaderID, title string, description *string, videoKey string, thumbKey *string, durationSec *int, premiumOnly bool) (*models.InstructionVideo, error) {
	video := &models.InstructionVideo{
		ID:         uuid.New().String(),
		MachineID:  machineID,
		UploaderID: uploaderID,
		Title:      title,
		VideoKey:   videoKey,
		PremiumOnly: premiumOnly,
	}

	var desc interface{}
	if description != nil {
		desc = *description
		copy := *description
		video.Description = &copy
	}

	var thumb interface{}
	if thumbKey != nil {
		thumb = *thumbKey
		copy := *thumbKey
		video.ThumbKey = &copy
	}

	var duration interface{}
	if durationSec != nil {
		duration = *durationSec
		copy := *durationSec
		video.DurationSec = &copy
	}

	query := `
		INSERT INTO instruction_videos (
			id, machine_id, uploader_id, title, description, video_key, thumb_key, duration_sec, premium_only
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at
	`

	if err := s.db.QueryRow(query, video.ID, machineID, uploaderID, title, desc, videoKey, thumb, duration, premiumOnly).Scan(&video.CreatedAt); err != nil {
		return nil, fmt.Errorf("failed to create video: %w", err)
	}

	return video, nil
}

// ListByMachine retrieves videos for a machine ordered by created_at DESC,id DESC.
func (s *Store) ListByMachine(machineID string, limit int, cursor *pagination.TimeDescCursor) (pagination.Paginated[models.InstructionVideo], error) {
	if limit <= 0 {
		return pagination.Paginated[models.InstructionVideo]{}, pagination.ErrInvalidLimit
	}

	baseQuery := `
        SELECT 
            iv.id,
            iv.machine_id,
            iv.uploader_id,
            iv.title,
            iv.description,
            iv.video_key,
            iv.thumb_key,
            iv.duration_sec,
            iv.premium_only,
            iv.likes_count,
            iv.created_at,
            u.name AS uploader_name,
            m.name AS machine_name
        FROM instruction_videos iv
        JOIN users u ON iv.uploader_id = u.id
        JOIN machines m ON iv.machine_id = m.id
        WHERE iv.machine_id = $1
	`

	var args []interface{}
	args = append(args, machineID)

	if cursor != nil {
		baseQuery += `
			AND (
				iv.created_at < $2 OR
				(iv.created_at = $2 AND iv.id < $3)
			)
		`
		args = append(args, cursor.CreatedAt, cursor.ID)
	}

    baseQuery += `
        ORDER BY iv.created_at DESC, iv.id DESC
    `

	limitPlaceholder := len(args) + 1
	baseQuery += fmt.Sprintf("\nLIMIT $%d", limitPlaceholder)
	args = append(args, limit+1)

	rows, err := s.db.Query(baseQuery, args...)
	if err != nil {
		return pagination.Paginated[models.InstructionVideo]{}, fmt.Errorf("failed to query videos: %w", err)
	}
	defer rows.Close()

	var items []models.InstructionVideo

	for rows.Next() {
    var (
        video        models.InstructionVideo
        desc         sql.NullString
        thumb        sql.NullString
        duration     sql.NullInt64
        uploaderName string
        machineName  string
        likeCount    int64
    )

        if err := rows.Scan(
            &video.ID,
            &video.MachineID,
            &video.UploaderID,
            &video.Title,
            &desc,
            &video.VideoKey,
            &thumb,
            &duration,
            &video.PremiumOnly,
            &likeCount,
            &video.CreatedAt,
            &uploaderName,
            &machineName,
        ); err != nil {
			return pagination.Paginated[models.InstructionVideo]{}, fmt.Errorf("failed to scan video: %w", err)
		}

		if desc.Valid {
			value := desc.String
			video.Description = &value
		}
		if thumb.Valid {
			value := thumb.String
			video.ThumbKey = &value
		}
        if duration.Valid {
            value := int(duration.Int64)
            video.DurationSec = &value
        }

        video.UploaderName = uploaderName
        video.MachineName = machineName
        video.LikeCount = int(likeCount)

		items = append(items, video)
	}

	if err := rows.Err(); err != nil {
		return pagination.Paginated[models.InstructionVideo]{}, fmt.Errorf("videos rows error: %w", err)
	}

	page, err := pagination.TimeDescPage(items, limit, func(item models.InstructionVideo) pagination.TimeDescCursor {
		return pagination.TimeDescCursor{
			CreatedAt: item.CreatedAt,
			ID:        item.ID,
		}
	})
	if err != nil {
		return pagination.Paginated[models.InstructionVideo]{}, err
	}

	return page, nil
}

// GetByID returns a video by ID or ErrVideoNotFound.
func (s *Store) GetByID(id string) (*models.InstructionVideo, error) {
	query := `
        SELECT 
            iv.id,
            iv.machine_id,
            iv.uploader_id,
            iv.title,
            iv.description,
            iv.video_key,
            iv.thumb_key,
            iv.duration_sec,
            iv.premium_only,
            iv.likes_count,
            iv.created_at,
            u.name AS uploader_name,
            m.name AS machine_name
        FROM instruction_videos iv
        JOIN users u ON iv.uploader_id = u.id
        JOIN machines m ON iv.machine_id = m.id
        WHERE iv.id = $1
	`

	var (
		video        models.InstructionVideo
		desc         sql.NullString
		thumb        sql.NullString
		duration     sql.NullInt64
		uploaderName string
		machineName  string
		likeCount    int64
	)

    err := s.db.QueryRow(query, id).Scan(
        &video.ID,
        &video.MachineID,
        &video.UploaderID,
        &video.Title,
        &desc,
        &video.VideoKey,
        &thumb,
        &duration,
        &video.PremiumOnly,
        &likeCount,
        &video.CreatedAt,
        &uploaderName,
        &machineName,
    )
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVideoNotFound
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	if desc.Valid {
		value := desc.String
		video.Description = &value
	}
	if thumb.Valid {
		value := thumb.String
		video.ThumbKey = &value
	}
	if duration.Valid {
		value := int(duration.Int64)
		video.DurationSec = &value
	}
    video.UploaderName = uploaderName
    video.MachineName = machineName
    video.LikeCount = int(likeCount)

	return &video, nil
}

// LikeVideo records a like for a video.
func (s *Store) LikeVideo(videoID, userID string) error {
	query := `
		WITH ins AS (
			INSERT INTO video_likes (video_id, user_id, created_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (video_id, user_id) DO NOTHING
			RETURNING 1
		)
		UPDATE instruction_videos
		SET likes_count = likes_count + (SELECT COUNT(*) FROM ins)
		WHERE id = $1
	`

	if _, err := s.db.Exec(query, videoID, userID); err != nil {
		return fmt.Errorf("failed to like video: %w", err)
	}

	return nil
}

// UnlikeVideo removes a like for a video.
func (s *Store) UnlikeVideo(videoID, userID string) error {
	query := `
		WITH deleted AS (
			DELETE FROM video_likes WHERE video_id = $1 AND user_id = $2 RETURNING 1
		)
		UPDATE instruction_videos
		SET likes_count = GREATEST(likes_count - (SELECT COUNT(*) FROM deleted), 0)
		WHERE id = $1
	`
	if _, err := s.db.Exec(query, videoID, userID); err != nil {
		return fmt.Errorf("failed to unlike video: %w", err)
	}
	return nil
}

// IsLiked returns whether the user liked the video.
func (s *Store) IsLiked(videoID, userID string) (bool, error) {
	var exists int
	err := s.db.QueryRow(`SELECT 1 FROM video_likes WHERE video_id = $1 AND user_id = $2`, videoID, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if video is liked: %w", err)
	}
	return true, nil
}

func (s *Store) ExportByUser(userID string) ([]models.InstructionVideo, error) {
	rows, err := s.db.Query(`
		SELECT id, machine_id, uploader_id, title, description, video_key, thumb_key, duration_sec, premium_only, likes_count, created_at
		FROM instruction_videos WHERE uploader_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.InstructionVideo
	for rows.Next() {
		var (
			video    models.InstructionVideo
			desc     sql.NullString
			thumb    sql.NullString
			duration sql.NullInt64
		)
		if err := rows.Scan(&video.ID, &video.MachineID, &video.UploaderID, &video.Title, &desc, &video.VideoKey, &thumb, &duration, &video.PremiumOnly, &video.LikeCount, &video.CreatedAt); err != nil {
			return nil, err
		}
		if desc.Valid {
			copy := desc.String
			video.Description = &copy
		}
		if thumb.Valid {
			copy := thumb.String
			video.ThumbKey = &copy
		}
		if duration.Valid {
			val := int(duration.Int64)
			video.DurationSec = &val
		}
		items = append(items, video)
	}
	return items, rows.Err()
}

func (s *Store) AnonymizeByUser(userID string) error {
	_, err := s.db.Exec(`
		UPDATE instruction_videos
		SET title = 'Deleted video', description = NULL, video_key = '', thumb_key = NULL, premium_only = FALSE
		WHERE uploader_id = $1
	`, userID)
	return err
}

func (s *Store) DeleteLikesByUser(userID string) error {
	_, err := s.db.Exec(`
		WITH removed AS (
			DELETE FROM video_likes WHERE user_id = $1 RETURNING video_id
		)
		UPDATE instruction_videos
		SET likes_count = GREATEST(likes_count - 1, 0)
		WHERE id IN (SELECT video_id FROM removed)
	`, userID)
	return err
}
