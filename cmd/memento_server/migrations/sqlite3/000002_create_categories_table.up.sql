-- Create categories table
CREATE TABLE categories (
    id BLOB PRIMARY KEY,
    user_id BLOB NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at DATETIME DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'now'))
);

-- Enforce uniqueness at the application level or using a unique index
CREATE UNIQUE INDEX uk_categories_user_name ON categories (user_id, name);
-- Index for user lookup
CREATE INDEX idx_categories_user_id ON categories (user_id);
