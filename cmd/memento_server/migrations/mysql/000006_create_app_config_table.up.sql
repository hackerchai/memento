-- Create app_config table: add id PK, make user_id unique non-FK
CREATE TABLE app_config (
    id BINARY(16) PRIMARY KEY,
    user_id BINARY(16) NOT NULL,
    scrape_img_offline BOOLEAN NOT NULL DEFAULT FALSE,
    llm_auto_gen_tags BOOLEAN NOT NULL DEFAULT FALSE,
    extract_links BOOLEAN NOT NULL DEFAULT FALSE,
    llm_profile_id BINARY(16),
    llm_provider VARCHAR(50),
    llm_auto_gen_abstract BOOLEAN NOT NULL DEFAULT FALSE,
    custom_user_agent TEXT,
    custom_scrape_timeout_seconds INT,
    custom_scrape_retry_times INT,
    custom_user_proxy TEXT,
    bypass_refer BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_app_config_user_id (user_id),
    INDEX idx_app_config_user_id_lookup (user_id) -- Separate index for non-unique lookups if needed, though UK covers it
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci; 