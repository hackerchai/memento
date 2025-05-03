-- Create tags table
CREATE TABLE tags (
    id BLOB PRIMARY KEY,
    user_id BLOB NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at DATETIME DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'now'))
);

CREATE UNIQUE INDEX uk_tags_user_name ON tags (user_id, name);
CREATE INDEX idx_tags_user_id ON tags (user_id);
