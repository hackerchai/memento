-- Add the slug column (allowing NULL)
ALTER TABLE tags ADD COLUMN slug TEXT;

-- Add unique index (application must ensure non-null slug)
-- Note: SQLite unique indexes allow multiple NULLs
CREATE UNIQUE INDEX uk_tags_user_slug ON tags (user_id, slug);
