-- Migration: Create check-ins and exercises tables
-- Created: 2024-01-01

-- Check-ins table
CREATE TABLE IF NOT EXISTS checkins (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, day)
);

-- Exercises table
CREATE TABLE IF NOT EXISTS exercises (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    gym_id UUID REFERENCES gyms(id) ON DELETE SET NULL,
    machine_id UUID REFERENCES machines(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Sets table
CREATE TABLE IF NOT EXISTS sets (
    id UUID PRIMARY KEY,
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    set_index INTEGER NOT NULL,
    reps INTEGER NOT NULL,
    weight_kg NUMERIC(6,2),
    rpe NUMERIC(3,1),
    notes TEXT
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_checkins_user_day ON checkins(user_id, day DESC);
CREATE INDEX IF NOT EXISTS idx_exercises_user_created ON exercises(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_exercises_gym ON exercises(gym_id);
CREATE INDEX IF NOT EXISTS idx_exercises_machine ON exercises(machine_id);
CREATE INDEX IF NOT EXISTS idx_sets_exercise ON sets(exercise_id);

-- Add comments
COMMENT ON TABLE checkins IS 'Daily check-ins for streak tracking';
COMMENT ON TABLE exercises IS 'Exercise sessions with optional gym/machine context';
COMMENT ON TABLE sets IS 'Individual sets within an exercise session';
