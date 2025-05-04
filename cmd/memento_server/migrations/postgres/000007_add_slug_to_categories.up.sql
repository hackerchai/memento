-- Add the slug column
ALTER TABLE categories ADD COLUMN slug VARCHAR(255);

-- Update existing rows to have a default slug (replace with actual slug generation if possible in SQL or handle in app)
-- Example: UPDATE categories SET slug = lower(regexp_replace(name, '[^a-zA-Z0-9]+', '-', 'g')) WHERE slug IS NULL;
-- For simplicity in migration, we might make it nullable first, backfill, then make NOT NULL.
-- Or add NOT NULL directly if the table is empty or backfill happens immediately after migration.

-- Make the column not null after potential backfill
ALTER TABLE categories ALTER COLUMN slug SET NOT NULL;

-- Add a unique constraint for user_id and slug
ALTER TABLE categories ADD CONSTRAINT categories_user_id_slug_key UNIQUE (user_id, slug);
