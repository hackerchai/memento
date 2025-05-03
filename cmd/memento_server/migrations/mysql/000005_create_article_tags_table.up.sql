-- Create article_tags join table (without foreign keys)
CREATE TABLE article_tags (
    article_id BINARY(16) NOT NULL,
    tag_id BINARY(16) NOT NULL,
    user_id BINARY(16) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (article_id, tag_id),
    INDEX idx_article_tags_article_id (article_id),
    INDEX idx_article_tags_tag_id (tag_id),
    INDEX idx_article_tags_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
