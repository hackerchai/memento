-- Remove the unique constraint
-- Note: Constraint name might differ slightly based on MySQL version/config, check if necessary
ALTER TABLE categories DROP KEY IF EXISTS uk_categories_user_slug;

-- Remove the slug column
ALTER TABLE categories DROP COLUMN IF EXISTS slug;
