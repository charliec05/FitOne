package models

import (
	"time"
)

// Checkin represents a daily check-in
type Checkin struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Day       time.Time `json:"day" db:"day"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CheckinStats represents streak statistics
type CheckinStats struct {
	CurrentStreakDays  int       `json:"current_streak_days"`
	LongestStreakDays  int       `json:"longest_streak_days"`
	LastCheckinDay     *time.Time `json:"last_checkin_day,omitempty"`
}

type LeaderboardEntry struct {
	UserID     string `json:"user_id"`
	UserName   string `json:"user_name"`
	StreakDays int    `json:"streak_days"`
}

// Exercise represents an exercise session
type Exercise struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	GymID     *string   `json:"gym_id,omitempty" db:"gym_id"`
	MachineID *string   `json:"machine_id,omitempty" db:"machine_id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	
	// Related data
	Sets []Set `json:"sets,omitempty"`
	
	// Display info
	GymName     *string `json:"gym_name,omitempty"`
	MachineName *string `json:"machine_name,omitempty"`
}

// Set represents a set within an exercise
type Set struct {
	ID        string    `json:"id" db:"id"`
	ExerciseID string   `json:"exercise_id" db:"exercise_id"`
	SetIndex  int       `json:"set_index" db:"set_index"`
	Reps      int       `json:"reps" db:"reps"`
	WeightKg  *float64  `json:"weight_kg,omitempty" db:"weight_kg"`
	RPE       *float64  `json:"rpe,omitempty" db:"rpe"`
	Notes     *string   `json:"notes,omitempty" db:"notes"`
}
