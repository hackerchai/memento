-- Drop the trigger
DROP TRIGGER IF EXISTS update_app_config_updated_at;

-- Drop the index
DROP INDEX IF EXISTS idx_app_config_user_id;

-- Drop the table
DROP TABLE IF EXISTS app_config; 