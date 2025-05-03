-- Drop the trigger
DROP TRIGGER IF EXISTS update_articles_updated_at ON articles;

-- Drop indexes
DROP INDEX IF EXISTS idx_articles_user_id;
DROP INDEX IF EXISTS idx_articles_category_id;
DROP INDEX IF EXISTS idx_articles_status;
DROP INDEX IF EXISTS idx_articles_created_at;

-- Drop the table
DROP TABLE IF EXISTS articles;
