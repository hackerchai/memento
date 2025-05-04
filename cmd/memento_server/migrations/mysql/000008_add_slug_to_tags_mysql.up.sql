-- Add the slug column
ALTER TABLE tags ADD COLUMN slug VARCHAR(255) NOT NULL;

-- Note: Add logic here or in application code to populate existing rows' slugs before making NOT NULL if table has data.

-- Add a unique constraint for user_id and slug
ALTER TABLE tags ADD CONSTRAINT uk_tags_user_slug UNIQUE (user_id, slug);
