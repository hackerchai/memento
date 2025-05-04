-- Drop the unique index
-- Note: Dropping the column itself requires complex table recreation
DROP INDEX IF EXISTS uk_tags_user_slug;