package models

import (
	"time"
)

// Gym represents a gym location
type Gym struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Lat       float64   `json:"lat" db:"lat"`
	Lng       float64   `json:"lng" db:"lng"`
	Address   string    `json:"address" db:"address"`
	Phone     *string   `json:"phone,omitempty" db:"phone"`
	Website   *string   `json:"website,omitempty" db:"website"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	
	// Computed fields
	DistanceKm *float64 `json:"distance_km,omitempty"`
	AvgRating   *float64 `json:"avg_rating,omitempty"`
	ReviewCount int      `json:"review_count"`
	MachineCount int    `json:"machine_count"`
}

// NearbyGym represents the minimal payload for the nearby gyms feed.
type NearbyGym struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Lat             float64  `json:"lat"`
	Lng             float64  `json:"lng"`
	Address         string   `json:"address"`
	DistanceM       float64  `json:"distance_m"`
	AvgRating       *float64 `json:"avg_rating,omitempty"`
	MachinesCount   int      `json:"machines_count"`
	PriceFromCents  *int     `json:"price_from_cents,omitempty"`
}

// GymPrice represents a gym membership price
type GymPrice struct {
	GymID      string `json:"gym_id" db:"gym_id"`
	PlanName   string `json:"plan_name" db:"plan_name"`
	PriceCents int    `json:"price_cents" db:"price_cents"`
	Period     string `json:"period" db:"period"`
}

// GymReview represents a user review for a gym
type GymReview struct {
	ID        string    `json:"id" db:"id"`
	GymID     string    `json:"gym_id" db:"gym_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Rating    int       `json:"rating" db:"rating"`
	Comment   *string   `json:"comment,omitempty" db:"comment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	
	// User info for display
	UserName string `json:"user_name,omitempty"`
}

// Machine represents a gym machine/equipment
type Machine struct {
	ID       string    `json:"id" db:"id"`
	Name     string    `json:"name" db:"name"`
	BodyPart string    `json:"body_part" db:"body_part"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GymMachine represents the relationship between gyms and machines
type GymMachine struct {
	GymID     string `json:"gym_id" db:"gym_id"`
	MachineID string `json:"machine_id" db:"machine_id"`
	Quantity  int    `json:"quantity" db:"quantity"`
}
