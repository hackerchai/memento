-- Drop the trigger first
DROP TRIGGER IF EXISTS update_categories_updated_at ON categories;

-- Drop the index
DROP INDEX IF EXISTS idx_categories_user_id;

-- Drop the table
DROP TABLE IF EXISTS categories;

-- Note: Do not drop the update_updated_at_column function here
-- as it might be used by other tables. It will be dropped
-- in the down migration of the last table that uses it (article_tags).
