-- Add locale column to app_config table
ALTER TABLE app_config
ADD COLUMN locale VARCHAR(10) NOT NULL DEFAULT 'en-US';

-- Update locale for the root user
UPDATE app_config
SET locale = 'en-US'
WHERE user_id = '00000000-0000-0000-0000-000000000001';
