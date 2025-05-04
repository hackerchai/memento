-- Remove the unique constraint
ALTER TABLE tags DROP KEY IF EXISTS uk_tags_user_slug;

-- Remove the slug column
ALTER TABLE tags DROP COLUMN IF EXISTS slug;
