-- Drop indexes
DROP INDEX IF EXISTS idx_article_tags_article_id;
DROP INDEX IF EXISTS idx_article_tags_tag_id;
DROP INDEX IF EXISTS idx_article_tags_user_id;

-- Drop the table
DROP TABLE IF EXISTS article_tags;
