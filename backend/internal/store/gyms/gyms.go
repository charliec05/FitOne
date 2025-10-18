package gyms

import (
	"database/sql"
	"fmt"
	"strings"

	"fitonex/backend/internal/models"
	"fitonex/backend/internal/pagination"

	"github.com/google/uuid"
)

// Store handles gym-related database operations
type Store struct {
	db *sql.DB
}

// New creates a new gyms store
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// GetNearby retrieves gyms within a specified radius using Haversine formula with pagination.
func (s *Store) GetNearby(lat, lng, radiusKm float64, limit int, cursor *pagination.DistanceAscCursor) (pagination.Paginated[models.NearbyGym], error) {
	if limit <= 0 {
		return pagination.Paginated[models.NearbyGym]{}, pagination.ErrInvalidLimit
	}

	radiusMeters := radiusKm * 1000

	query := `
WITH distance_base AS (
	SELECT 
		g.id,
		g.name,
		g.lat,
		g.lng,
		g.address,
		1000 * 6371 * acos(
			LEAST(
				1,
				GREATEST(
					-1,
					cos(radians($1)) * cos(radians(g.lat)) * cos(radians(g.lng) - radians($2)) +
					sin(radians($1)) * sin(radians(g.lat))
				)
			)
		) AS distance_m
	FROM gyms g
),
review_stats AS (
	SELECT gym_id, AVG(rating)::float AS avg_rating
	FROM gym_reviews
	GROUP BY gym_id
),
machine_stats AS (
	SELECT gym_id, COUNT(*)::int AS machines_count
	FROM gym_machines
	GROUP BY gym_id
),
price_stats AS (
	SELECT gym_id, MIN(price_cents) AS price_from_cents
	FROM gym_prices
	GROUP BY gym_id
)
SELECT 
	b.id,
	b.name,
	b.lat,
	b.lng,
	b.address,
	b.distance_m,
	r.avg_rating,
	COALESCE(m.machines_count, 0) AS machines_count,
	COALESCE(pc.price_from_cents, p.price_from_cents) AS price_from_cents
FROM distance_base b
LEFT JOIN review_stats r ON r.gym_id = b.id
LEFT JOIN machine_stats m ON m.gym_id = b.id
LEFT JOIN price_stats p ON p.gym_id = b.id
LEFT JOIN gym_price_cache pc ON pc.gym_id = b.id
WHERE b.distance_m <= $3
`

	args := []any{lat, lng, radiusMeters, limit + 1}
	if cursor != nil {
		query += `
	AND (
		b.distance_m > $5 OR (b.distance_m = $5 AND b.id > $6)
	)`
		args = append(args, cursor.DistanceM, cursor.ID)
	}

	query += `
ORDER BY b.distance_m ASC, b.id ASC
LIMIT $4
`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return pagination.Paginated[models.NearbyGym]{}, fmt.Errorf("failed to query nearby gyms: %w", err)
	}
	defer rows.Close()

	var (
		results []models.NearbyGym
	)

	for rows.Next() {
		var (
			item           models.NearbyGym
			avgRating      sql.NullFloat64
			priceFromCents sql.NullInt64
		)

		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Lat,
			&item.Lng,
			&item.Address,
			&item.DistanceM,
			&avgRating,
			&item.MachinesCount,
			&priceFromCents,
		); err != nil {
			return pagination.Paginated[models.NearbyGym]{}, fmt.Errorf("failed to scan nearby gym: %w", err)
		}

		if avgRating.Valid {
			value := avgRating.Float64
			item.AvgRating = &value
		}
		if priceFromCents.Valid {
			value := int(priceFromCents.Int64)
			item.PriceFromCents = &value
		}

		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return pagination.Paginated[models.NearbyGym]{}, fmt.Errorf("nearby gyms rows error: %w", err)
	}

	page, err := pagination.DistanceAscPage(results, limit, func(item models.NearbyGym) pagination.DistanceAscCursor {
		return pagination.DistanceAscCursor{
			DistanceM: item.DistanceM,
			ID:        item.ID,
		}
	})
	if err != nil {
		return pagination.Paginated[models.NearbyGym]{}, err
	}

	return page, nil
}

// GetByID retrieves a gym by ID
func (s *Store) GetByID(id string) (*models.Gym, error) {
	query := `
	SELECT 
		g.id, g.name, g.lat, g.lng, g.address, g.phone, g.website, g.created_at,
		COALESCE(AVG(gr.rating), 0) as avg_rating,
		COUNT(gr.id) as review_count,
		COUNT(DISTINCT gm.machine_id) as machine_count,
		COALESCE(pc.price_from_cents,
			(SELECT MIN(price_cents) FROM gym_prices WHERE gym_id = g.id)
		) AS price_from_cents
	FROM gyms g
	LEFT JOIN gym_reviews gr ON g.id = gr.gym_id
	LEFT JOIN gym_machines gm ON g.id = gm.gym_id
	LEFT JOIN gym_price_cache pc ON pc.gym_id = g.id
	WHERE g.id = $1
	GROUP BY g.id, g.name, g.lat, g.lng, g.address, g.phone, g.website, g.created_at, pc.price_from_cents
`

	var gym models.Gym
	var avgRating float64
	var reviewCount, machineCount int
	var priceFrom sql.NullInt64

	err := s.db.QueryRow(query, id).Scan(
		&gym.ID, &gym.Name, &gym.Lat, &gym.Lng, &gym.Address,
		&gym.Phone, &gym.Website, &gym.CreatedAt,
		&avgRating, &reviewCount, &machineCount, &priceFrom,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("gym not found")
		}
		return nil, fmt.Errorf("failed to get gym: %w", err)
	}

if avgRating > 0 {
	gym.AvgRating = &avgRating
}
gym.ReviewCount = reviewCount
gym.MachineCount = machineCount
	if priceFrom.Valid {
		value := int(priceFrom.Int64)
		gym.PriceFromCents = &value
	}

	return &gym, nil
}

// GetMachines retrieves machines for a specific gym
func (s *Store) GetMachines(gymID string) ([]models.Machine, error) {
	query := `
		SELECT m.id, m.name, m.body_part, m.created_at, gm.quantity
		FROM machines m
		JOIN gym_machines gm ON m.id = gm.machine_id
		WHERE gm.gym_id = $1
		ORDER BY m.name
	`

	rows, err := s.db.Query(query, gymID)
	if err != nil {
		return nil, fmt.Errorf("failed to query gym machines: %w", err)
	}
	defer rows.Close()

	var machines []models.Machine
	for rows.Next() {
		var machine models.Machine
		var quantity int

		err := rows.Scan(
			&machine.ID, &machine.Name, &machine.BodyPart, &machine.CreatedAt, &quantity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan machine: %w", err)
		}

		machines = append(machines, machine)
	}

	return machines, nil
}

// GetPrices retrieves pricing plans for a specific gym
func (s *Store) GetPrices(gymID string) ([]models.GymPrice, error) {
	query := `
		SELECT gym_id, plan_name, price_cents, period
		FROM gym_prices
		WHERE gym_id = $1
		ORDER BY price_cents ASC
	`

	rows, err := s.db.Query(query, gymID)
	if err != nil {
		return nil, fmt.Errorf("failed to query gym prices: %w", err)
	}
	defer rows.Close()

	var prices []models.GymPrice
	for rows.Next() {
		var price models.GymPrice

		err := rows.Scan(
			&price.GymID, &price.PlanName, &price.PriceCents, &price.Period,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan price: %w", err)
		}

		prices = append(prices, price)
	}

	return prices, nil
}

// CreateReview creates a new gym review
func (s *Store) CreateReview(gymID, userID string, rating int, comment string) (*models.GymReview, error) {
	review := &models.GymReview{
		ID:      uuid.New().String(),
		GymID:   gymID,
		UserID:  userID,
		Rating:  rating,
		Comment: &comment,
	}

	query := `
		INSERT INTO gym_reviews (id, gym_id, user_id, rating, comment)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`

	err := s.db.QueryRow(query, review.ID, review.GymID, review.UserID, review.Rating, review.Comment).Scan(&review.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	return review, nil
}

// GetReviews retrieves reviews for a gym with pagination
func (s *Store) GetReviews(gymID, cursor string, limit int) ([]models.GymReview, string, error) {
	var query string
	var args []interface{}

	if cursor == "" {
		query = `
			SELECT gr.id, gr.gym_id, gr.user_id, gr.rating, gr.comment, gr.created_at, u.name
			FROM gym_reviews gr
			JOIN users u ON gr.user_id = u.id
			WHERE gr.gym_id = $1
			ORDER BY gr.created_at DESC
			LIMIT $2
		`
		args = []interface{}{gymID, limit + 1}
	} else {
		query = `
			SELECT gr.id, gr.gym_id, gr.user_id, gr.rating, gr.comment, gr.created_at, u.name
			FROM gym_reviews gr
			JOIN users u ON gr.user_id = u.id
			WHERE gr.gym_id = $1 AND gr.created_at < $2
			ORDER BY gr.created_at DESC
			LIMIT $3
		`
		args = []interface{}{gymID, cursor, limit + 1}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("failed to query reviews: %w", err)
	}
	defer rows.Close()

	var reviews []models.GymReview
	var nextCursor string

	for rows.Next() {
		var review models.GymReview
		var userName string

		err := rows.Scan(
			&review.ID, &review.GymID, &review.UserID, &review.Rating, 
			&review.Comment, &review.CreatedAt, &userName,
		)
		if err != nil {
			return nil, "", fmt.Errorf("failed to scan review: %w", err)
		}

		review.UserName = userName
		reviews = append(reviews, review)
	}

	// Check if there are more results
	if len(reviews) > limit {
		nextCursor = reviews[limit-1].CreatedAt.Format("2006-01-02T15:04:05.999999Z07:00")
		reviews = reviews[:limit]
	}

	return reviews, nextCursor, nil
}

func (s *Store) Search(name string, limit int, cursor *pagination.ScoreDescCursor, prefix bool) (pagination.Paginated[models.GymSearchResult], error) {
	if limit <= 0 {
		return pagination.Paginated[models.GymSearchResult]{}, pagination.ErrInvalidLimit
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return pagination.Paginated[models.GymSearchResult]{}, nil
	}

	var (
		query string
		args  []any
	)

	if prefix {
		query = `
SELECT id, name, address, 1.0 AS score
FROM gyms
WHERE LOWER(name) LIKE LOWER($1) || '%'
`
		args = append(args, name)
		if cursor != nil {
			query += ` AND id > $3`
			args = append(args, limit+1, cursor.ID)
		} else {
			args = append(args, limit+1)
		}
		query += " ORDER BY score DESC, name ASC, id ASC LIMIT $2"
	} else {
		query = `
WITH ranked AS (
    SELECT id, name, address, similarity(name, $1) AS score
    FROM gyms
)
SELECT id, name, address, score
FROM ranked
WHERE score > 0.1
`
		args = append(args, name, limit+1)
		if cursor != nil {
			query += ` AND (score < $3 OR (score = $3 AND id > $4))`
			args = append(args, cursor.Score, cursor.ID)
		}
		query += " ORDER BY score DESC, id ASC LIMIT $2"
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return pagination.Paginated[models.GymSearchResult]{}, fmt.Errorf("search gyms: %w", err)
	}
	defer rows.Close()

	var results []models.GymSearchResult
	for rows.Next() {
		var item models.GymSearchResult
		if err := rows.Scan(&item.ID, &item.Name, &item.Address, &item.Score); err != nil {
			return pagination.Paginated[models.GymSearchResult]{}, fmt.Errorf("scan gym search: %w", err)
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return pagination.Paginated[models.GymSearchResult]{}, err
	}

	page, err := pagination.ScoreDescPage(results, limit, func(item models.GymSearchResult) pagination.ScoreDescCursor {
		return pagination.ScoreDescCursor{
			Score: item.Score,
			ID:    item.ID,
		}
	})
	if err != nil {
		return pagination.Paginated[models.GymSearchResult]{}, err
	}
	return page, nil
}
