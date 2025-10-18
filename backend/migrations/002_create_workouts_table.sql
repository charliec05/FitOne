-- Migration: Create workouts table
-- Created: 2024-01-01

CREATE TABLE IF NOT EXISTS workouts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    duration INTEGER NOT NULL,
    type VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_workouts_user_id ON workouts(user_id);
CREATE INDEX IF NOT EXISTS idx_workouts_created_at ON workouts(created_at);
CREATE INDEX IF NOT EXISTS idx_workouts_user_created ON workouts(user_id, created_at);

-- Add comments
COMMENT ON TABLE workouts IS 'User workouts for the FitONEX application';
COMMENT ON COLUMN workouts.id IS 'Unique identifier for the workout';
COMMENT ON COLUMN workouts.user_id IS 'Reference to the user who owns this workout';
COMMENT ON COLUMN workouts.name IS 'Workout name';
COMMENT ON COLUMN workouts.description IS 'Optional workout description';
COMMENT ON COLUMN workouts.duration IS 'Workout duration in minutes';
COMMENT ON COLUMN workouts.type IS 'Type of workout (e.g., cardio, strength, yoga)';
COMMENT ON COLUMN workouts.created_at IS 'Workout creation timestamp';
COMMENT ON COLUMN workouts.updated_at IS 'Last update timestamp';
