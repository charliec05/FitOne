package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Password  string    `json:"-" db:"password"` // Hidden from JSON
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	PremiumUntil  *time.Time `json:"premium_until,omitempty" db:"premium_until"`
	OAuthProvider *string    `json:"oauth_provider,omitempty" db:"oauth_provider"`
	OAuthID       *string    `json:"oauth_id,omitempty" db:"oauth_id"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

func (u *User) IsPremium() bool {
	if u == nil || u.PremiumUntil == nil {
		return false
	}
	return u.PremiumUntil.After(time.Now())
}
