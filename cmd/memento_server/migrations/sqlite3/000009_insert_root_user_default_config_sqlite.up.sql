-- Insert default config for root user
INSERT INTO app_config (
    id, -- Provide ID for SQLite
    user_id,
    scrape_img_offline,
    llm_auto_gen_tags,
    extract_links,
    llm_auto_gen_abstract,
    bypass_refer
    -- created_at, updated_at will use defaults
    -- other nullable fields default to NULL
) VALUES (
    randomblob(16), -- Generate ID
    (SELECT x'00000000000000000000000000000001'), -- Root User ID as BLOB
    1,      -- scrape_img_offline (TRUE)
    0,      -- llm_auto_gen_tags (FALSE)
    0,      -- extract_links (FALSE)
    0,      -- llm_auto_gen_abstract (FALSE)
    0       -- bypass_refer (FALSE)
);
