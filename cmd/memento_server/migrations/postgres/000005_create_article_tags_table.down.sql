-- Drop indexes
DROP INDEX IF EXISTS idx_article_tags_article_id;
DROP INDEX IF EXISTS idx_article_tags_tag_id;
DROP INDEX IF EXISTS idx_article_tags_user_id;

-- Drop the table
DROP TABLE IF EXISTS article_tags;

-- Finally, drop the shared update function as it's no longer needed by any table
DROP FUNCTION IF EXISTS update_updated_at_column();
