-- Insert default config for root user
INSERT INTO app_config (
    user_id,
    scrape_img_offline,
    llm_auto_gen_tags,
    extract_links,
    llm_auto_gen_abstract,
    bypass_refer
    -- id, created_at, updated_at will use defaults
    -- other nullable fields default to NULL
) VALUES (
    '00000000-0000-0000-0000-000000000001', -- Root User ID
    TRUE,   -- scrape_img_offline
    FALSE,  -- llm_auto_gen_tags
    FALSE,  -- extract_links
    FALSE,  -- llm_auto_gen_abstract
    FALSE   -- bypass_refer
);
