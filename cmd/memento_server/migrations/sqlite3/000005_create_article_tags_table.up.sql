-- Create article_tags join table
CREATE TABLE article_tags (
    article_id BLOB NOT NULL,
    tag_id BLOB NOT NULL,
    user_id BLOB NOT NULL,
    created_at DATETIME DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'now')),
    PRIMARY KEY (article_id, tag_id)
);

-- Add indexes for efficient lookups
CREATE INDEX idx_article_tags_article_id ON article_tags(article_id);
CREATE INDEX idx_article_tags_tag_id ON article_tags(tag_id);
CREATE INDEX idx_article_tags_user_id ON article_tags(user_id);
