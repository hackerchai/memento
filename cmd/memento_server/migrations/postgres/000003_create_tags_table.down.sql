-- Drop the trigger
DROP TRIGGER IF EXISTS update_tags_updated_at ON tags;

-- Drop the index
DROP INDEX IF EXISTS idx_tags_user_id;

-- Drop the table
DROP TABLE IF EXISTS tags;
