-- Remove the unique constraint
ALTER TABLE tags DROP CONSTRAINT IF EXISTS tags_user_id_slug_key;

-- Remove the slug column
ALTER TABLE tags DROP COLUMN IF EXISTS slug;
