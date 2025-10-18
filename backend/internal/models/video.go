package models

import (
	"time"
)

// InstructionVideo represents an instruction video for a machine
type InstructionVideo struct {
	ID          string    `json:"id" db:"id"`
	MachineID   string    `json:"machine_id" db:"machine_id"`
	UploaderID string    `json:"uploader_id" db:"uploader_id"`
	Title       string    `json:"title" db:"title"`
	Description *string   `json:"description,omitempty" db:"description"`
	VideoKey    string    `json:"video_key" db:"video_key"`
	ThumbKey    *string   `json:"thumb_key,omitempty" db:"thumb_key"`
	DurationSec *int      `json:"duration_sec,omitempty" db:"duration_sec"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	
	// Computed fields
	LikeCount int `json:"like_count"`
	IsLiked   bool `json:"is_liked"`
	PlayURL   string `json:"play_url,omitempty"`
	
	// User info for display
	UploaderName string `json:"uploader_name,omitempty"`
	MachineName  string `json:"machine_name,omitempty"`
}

// VideoLike represents a user's like for a video
type VideoLike struct {
	VideoID   string    `json:"video_id" db:"video_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
