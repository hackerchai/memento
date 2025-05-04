-- Create app_config table: add id PK, make user_id unique non-FK
CREATE TABLE app_config (
    id BLOB PRIMARY KEY,
    user_id BLOB NOT NULL UNIQUE,
    scrape_img_offline INTEGER NOT NULL DEFAULT 0,
    llm_auto_gen_tags INTEGER NOT NULL DEFAULT 0,
    extract_links INTEGER NOT NULL DEFAULT 0,
    llm_profile_id BLOB,
    llm_provider TEXT,
    llm_auto_gen_abstract INTEGER NOT NULL DEFAULT 0,
    custom_user_agent TEXT,
    custom_scrape_timeout_seconds INTEGER,
    custom_scrape_retry_times INTEGER,
    custom_user_proxy TEXT,
    bypass_refer INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at DATETIME DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'now'))
);

-- Index for user lookup (unique constraint also creates an index, but explicit doesn't hurt)
CREATE INDEX idx_app_config_user_id ON app_config(user_id);

-- Add trigger to update updated_at timestamp
CREATE TRIGGER update_app_config_updated_at
AFTER UPDATE ON app_config
FOR EACH ROW
BEGIN
    UPDATE app_config SET updated_at = STRFTIME('%Y-%m-%d %H:%M:%f', 'now') WHERE id = OLD.id;
END; 