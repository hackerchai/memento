-- Create articles table
CREATE TABLE articles (
    id BLOB PRIMARY KEY,
    user_id BLOB NOT NULL,
    category_id BLOB,
    title TEXT NOT NULL,
    html TEXT,
    author TEXT,
    description TEXT,
    llm_description TEXT,
    plain_text TEXT,
    og_image_url TEXT,
    url TEXT NOT NULL,
    is_offline INTEGER NOT NULL DEFAULT 0, -- Use INTEGER for BOOLEAN (0=false, 1=true)
    status INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at DATETIME DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'now'))
);

CREATE UNIQUE INDEX uk_articles_user_url ON articles (user_id, url);
CREATE INDEX idx_articles_user_id ON articles (user_id);
CREATE INDEX idx_articles_category_id ON articles (category_id);
CREATE INDEX idx_articles_status ON articles (status);
CREATE INDEX idx_articles_created_at ON articles (created_at);
