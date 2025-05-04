-- Add the slug column
ALTER TABLE tags ADD COLUMN slug VARCHAR(255);

-- Update existing rows (similar considerations as categories)
-- Example: UPDATE tags SET slug = lower(regexp_replace(name, '[^a-zA-Z0-9]+', '-', 'g')) WHERE slug IS NULL;

-- Make the column not null after potential backfill
ALTER TABLE tags ALTER COLUMN slug SET NOT NULL;

-- Add a unique constraint for user_id and slug
ALTER TABLE tags ADD CONSTRAINT tags_user_id_slug_key UNIQUE (user_id, slug);
