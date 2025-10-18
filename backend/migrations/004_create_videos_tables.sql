-- Migration: Create instruction videos tables
-- Created: 2024-01-01

-- Instruction videos table
CREATE TABLE IF NOT EXISTS instruction_videos (
    id UUID PRIMARY KEY,
    machine_id UUID NOT NULL REFERENCES machines(id) ON DELETE CASCADE,
    uploader_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT,
    video_key TEXT NOT NULL,
    thumb_key TEXT,
    duration_sec INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Video likes table
CREATE TABLE IF NOT EXISTS video_likes (
    video_id UUID NOT NULL REFERENCES instruction_videos(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (video_id, user_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_instruction_videos_machine ON instruction_videos(machine_id);
CREATE INDEX IF NOT EXISTS idx_instruction_videos_uploader ON instruction_videos(uploader_id);
CREATE INDEX IF NOT EXISTS idx_instruction_videos_created ON instruction_videos(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_video_likes_video ON video_likes(video_id);
CREATE INDEX IF NOT EXISTS idx_video_likes_user ON video_likes(user_id);

-- Add comments
COMMENT ON TABLE instruction_videos IS 'Instruction videos for gym machines';
COMMENT ON TABLE video_likes IS 'User likes for instruction videos';
