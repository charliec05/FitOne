package migrations

import (
	"database/sql"
	"fmt"
)

// Up runs all pending migrations
func Up(db *sql.DB) error {
	statements := []string{
		"CREATE EXTENSION IF NOT EXISTS pg_trgm",
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			name VARCHAR(255) NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS workouts (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			duration INTEGER NOT NULL,
			type VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS gyms (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			lat DOUBLE PRECISION NOT NULL,
			lng DOUBLE PRECISION NOT NULL,
			address TEXT NOT NULL,
			phone TEXT,
			website TEXT,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS gym_prices (
			gym_id UUID NOT NULL REFERENCES gyms(id) ON DELETE CASCADE,
			plan_name TEXT NOT NULL,
			price_cents INTEGER NOT NULL,
			period TEXT NOT NULL,
			PRIMARY KEY (gym_id, plan_name)
		)`,
		`CREATE TABLE IF NOT EXISTS machines (
			id UUID PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			body_part TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS gym_machines (
			gym_id UUID NOT NULL REFERENCES gyms(id) ON DELETE CASCADE,
			machine_id UUID NOT NULL REFERENCES machines(id) ON DELETE CASCADE,
			quantity INTEGER NOT NULL DEFAULT 1,
			PRIMARY KEY (gym_id, machine_id)
		)`,
		`CREATE TABLE IF NOT EXISTS gym_reviews (
			id UUID PRIMARY KEY,
			gym_id UUID NOT NULL REFERENCES gyms(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
			comment TEXT,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS instruction_videos (
			id UUID PRIMARY KEY,
			machine_id UUID NOT NULL REFERENCES machines(id) ON DELETE CASCADE,
			uploader_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title TEXT NOT NULL,
			description TEXT,
			video_key TEXT NOT NULL,
			thumb_key TEXT,
			duration_sec INTEGER,
			premium_only BOOLEAN NOT NULL DEFAULT FALSE,
			likes_count INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS video_likes (
			video_id UUID NOT NULL REFERENCES instruction_videos(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			PRIMARY KEY (video_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS checkins (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			day DATE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			UNIQUE (user_id, day)
		)`,
		`CREATE TABLE IF NOT EXISTS exercises (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			gym_id UUID REFERENCES gyms(id) ON DELETE SET NULL,
			machine_id UUID REFERENCES machines(id) ON DELETE SET NULL,
			name TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS sets (
			id UUID PRIMARY KEY,
			exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
			set_index INTEGER NOT NULL,
			reps INTEGER NOT NULL,
			weight_kg NUMERIC(6,2),
			rpe NUMERIC(3,1),
			notes TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS moderation_reports (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			object_type TEXT NOT NULL,
			object_id UUID NOT NULL,
			reason TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS gym_price_cache (
			gym_id UUID PRIMARY KEY REFERENCES gyms(id) ON DELETE CASCADE,
			price_from_cents INTEGER NOT NULL,
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_workouts_user_id ON workouts(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_workouts_user_created ON workouts(user_id, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_gyms_location ON gyms(lat, lng)",
		"CREATE INDEX IF NOT EXISTS idx_gym_reviews_gym_created ON gym_reviews(gym_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_video_likes_video ON video_likes(video_id)",
		"CREATE INDEX IF NOT EXISTS idx_checkins_user_day ON checkins(user_id, day DESC)",
		"CREATE INDEX IF NOT EXISTS idx_exercises_user_created ON exercises(user_id, created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_sets_exercise ON sets(exercise_id)",
		"CREATE INDEX IF NOT EXISTS idx_gyms_name_trgm ON gyms USING gin (name gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_machines_name_trgm ON machines USING gin (name gin_trgm_ops)",
		"CREATE INDEX IF NOT EXISTS idx_machines_body_part_trgm ON machines USING gin (body_part gin_trgm_ops)",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS premium_until TIMESTAMP WITH TIME ZONE",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_provider TEXT",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS oauth_id TEXT",
		"ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_users_oauth ON users(oauth_provider, oauth_id)",
		"ALTER TABLE instruction_videos ADD COLUMN IF NOT EXISTS likes_count INTEGER NOT NULL DEFAULT 0",
		"ALTER TABLE instruction_videos ADD COLUMN IF NOT EXISTS premium_only BOOLEAN NOT NULL DEFAULT FALSE",
		`CREATE TABLE IF NOT EXISTS video_comments (
			id UUID PRIMARY KEY,
			video_id UUID NOT NULL REFERENCES instruction_videos(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			comment TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		"CREATE INDEX IF NOT EXISTS idx_video_comments_video_created ON video_comments(video_id, created_at DESC)",
		`CREATE TABLE IF NOT EXISTS password_resets (
			token TEXT PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)`,
		"CREATE INDEX IF NOT EXISTS idx_password_resets_user ON password_resets(user_id)",
}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed on %q: %w", stmt, err)
		}
	}

	return nil
}

// Down rolls back the last migration
func Down(db *sql.DB) error {
	tables := []string{
		"DROP TABLE IF EXISTS gym_price_cache",
		"DROP TABLE IF EXISTS moderation_reports",
		"DROP TABLE IF EXISTS sets",
		"DROP TABLE IF EXISTS exercises",
		"DROP TABLE IF EXISTS checkins",
		"DROP TABLE IF EXISTS video_likes",
		"DROP TABLE IF EXISTS instruction_videos",
		"DROP TABLE IF EXISTS gym_reviews",
		"DROP TABLE IF EXISTS gym_machines",
		"DROP TABLE IF EXISTS machines",
		"DROP TABLE IF EXISTS gym_prices",
		"DROP TABLE IF EXISTS gyms",
		"DROP TABLE IF EXISTS workouts",
		"DROP TABLE IF EXISTS users",
	}

	for _, stmt := range tables {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to drop table: %w", err)
		}
	}
	return nil
}
