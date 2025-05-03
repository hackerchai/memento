-- Create articles table
CREATE TABLE articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL, -- Allow articles without category or keep category if deleted? SET NULL chosen here.
    title TEXT NOT NULL,
    html TEXT,
    author TEXT,
    description TEXT, -- Renamed from 'desc' to avoid potential SQL keyword conflict
    llm_description TEXT, -- Renamed from 'llm_desc'
    plain_text TEXT, -- Renamed from 'text' for clarity
    og_image_url TEXT,
    url TEXT NOT NULL,
    is_offline BOOLEAN NOT NULL DEFAULT FALSE,
    status SMALLINT NOT NULL DEFAULT 0, -- Use SMALLINT for status codes
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, url) -- Ensure URL is unique per user
);

-- Add indexes
CREATE INDEX idx_articles_user_id ON articles(user_id);
CREATE INDEX idx_articles_category_id ON articles(category_id);
CREATE INDEX idx_articles_status ON articles(status);
CREATE INDEX idx_articles_created_at ON articles(created_at); -- Index for sorting by creation time

-- Trigger to update updated_at on articles table update
-- Reusing the function created in the categories migration
CREATE TRIGGER update_articles_updated_at
BEFORE UPDATE ON articles
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
