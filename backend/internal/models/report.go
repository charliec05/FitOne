package models

import "time"

type ModerationReport struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	ObjectType string    `json:"object_type"`
	ObjectID   string    `json:"object_id"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}
