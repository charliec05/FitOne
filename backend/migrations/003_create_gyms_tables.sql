-- Migration: Create gyms and related tables
-- Created: 2024-01-01

-- Gyms table
CREATE TABLE IF NOT EXISTS gyms (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    lat DOUBLE PRECISION NOT NULL,
    lng DOUBLE PRECISION NOT NULL,
    address TEXT NOT NULL,
    phone TEXT,
    website TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Gym prices table
CREATE TABLE IF NOT EXISTS gym_prices (
    gym_id UUID NOT NULL REFERENCES gyms(id) ON DELETE CASCADE,
    plan_name TEXT NOT NULL,
    price_cents INTEGER NOT NULL,
    period TEXT NOT NULL,
    PRIMARY KEY (gym_id, plan_name)
);

-- Machines table
CREATE TABLE IF NOT EXISTS machines (
    id UUID PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    body_part TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Gym machines junction table
CREATE TABLE IF NOT EXISTS gym_machines (
    gym_id UUID NOT NULL REFERENCES gyms(id) ON DELETE CASCADE,
    machine_id UUID NOT NULL REFERENCES machines(id) ON DELETE CASCADE,
    quantity INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (gym_id, machine_id)
);

-- Gym reviews table
CREATE TABLE IF NOT EXISTS gym_reviews (
    id UUID PRIMARY KEY,
    gym_id UUID NOT NULL REFERENCES gyms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_gyms_location ON gyms(lat, lng);
CREATE INDEX IF NOT EXISTS idx_gym_reviews_gym_created ON gym_reviews(gym_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_gym_reviews_user ON gym_reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_machines_body_part ON machines(body_part);
CREATE INDEX IF NOT EXISTS idx_gym_machines_gym ON gym_machines(gym_id);
CREATE INDEX IF NOT EXISTS idx_gym_machines_machine ON gym_machines(machine_id);

-- Add comments
COMMENT ON TABLE gyms IS 'Gym locations with coordinates and contact info';
COMMENT ON TABLE gym_prices IS 'Gym membership pricing plans';
COMMENT ON TABLE machines IS 'Available gym machines/equipment';
COMMENT ON TABLE gym_machines IS 'Junction table linking gyms to their available machines';
COMMENT ON TABLE gym_reviews IS 'User reviews and ratings for gyms';
