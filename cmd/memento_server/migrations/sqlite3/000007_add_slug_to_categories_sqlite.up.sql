-- Add the slug column (allowing NULL)
ALTER TABLE categories ADD COLUMN slug TEXT;

-- Add unique index (application must ensure non-null slug)
-- Note: SQLite unique indexes allow multiple NULLs
CREATE UNIQUE INDEX uk_categories_user_slug ON categories (user_id, slug);