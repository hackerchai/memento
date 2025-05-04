-- Create app_config table for user-specific settings
CREATE TABLE app_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    scrape_img_offline BOOLEAN NOT NULL DEFAULT FALSE,
    llm_auto_gen_tags BOOLEAN NOT NULL DEFAULT FALSE,
    extract_links BOOLEAN NOT NULL DEFAULT FALSE,
    llm_profile_id UUID, -- Foreign key constraint can be added later if LLM profiles table exists
    llm_provider VARCHAR(50),
    llm_auto_gen_abstract BOOLEAN NOT NULL DEFAULT FALSE,
    custom_user_agent TEXT,
    custom_scrape_timeout_seconds INTEGER,
    custom_scrape_retry_times INTEGER,
    custom_user_proxy TEXT,
    bypass_refer BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add index on user_id for faster lookups
CREATE INDEX idx_app_config_user_id ON app_config(user_id);

-- Trigger to update updated_at on app_config table update
-- Reusing the function created in previous migrations
CREATE TRIGGER update_app_config_updated_at
BEFORE UPDATE ON app_config
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column(); 