-- Remove the unique constraint
ALTER TABLE categories DROP CONSTRAINT IF EXISTS categories_user_id_slug_key;

-- Remove the slug column
ALTER TABLE categories DROP COLUMN IF EXISTS slug;
