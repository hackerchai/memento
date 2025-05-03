-- Drop indexes
DROP INDEX IF EXISTS uk_categories_user_name;
DROP INDEX IF EXISTS idx_categories_user_id;

-- Drop the table
DROP TABLE IF EXISTS categories;
