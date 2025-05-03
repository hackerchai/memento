-- Create articles table (without foreign keys)
CREATE TABLE articles (
    id BINARY(16) PRIMARY KEY,
    user_id BINARY(16) NOT NULL,
    category_id BINARY(16),
    title TEXT NOT NULL,
    html LONGTEXT,
    author VARCHAR(255),
    description TEXT,
    llm_description TEXT,
    plain_text LONGTEXT,
    og_image_url VARCHAR(2048), -- Using VARCHAR for potentially shorter URLs and easier indexing if needed
    url VARCHAR(2048) NOT NULL, -- Use VARCHAR for URL to allow unique constraint with reasonable length
    is_offline BOOLEAN NOT NULL DEFAULT FALSE,
    status SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_articles_user_url (user_id, url(255)), -- Unique constraint on user and partial URL
    INDEX idx_articles_user_id (user_id),
    INDEX idx_articles_category_id (category_id),
    INDEX idx_articles_status (status),
    INDEX idx_articles_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
