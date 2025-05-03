-- Drop indexes
DROP INDEX IF EXISTS uk_tags_user_name;
DROP INDEX IF EXISTS idx_tags_user_id;

-- Drop the table
DROP TABLE IF EXISTS tags;
