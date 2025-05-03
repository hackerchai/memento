-- Drop indexes
DROP INDEX IF EXISTS uk_articles_user_url;
DROP INDEX IF EXISTS idx_articles_user_id;
DROP INDEX IF EXISTS idx_articles_category_id;
DROP INDEX IF EXISTS idx_articles_status;
DROP INDEX IF EXISTS idx_articles_created_at;

-- Drop the table
DROP TABLE IF EXISTS articles;
