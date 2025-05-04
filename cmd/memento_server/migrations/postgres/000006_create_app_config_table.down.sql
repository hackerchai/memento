-- Drop the trigger
DROP TRIGGER IF EXISTS update_app_config_updated_at ON app_config;

-- Drop the index
DROP INDEX IF EXISTS idx_app_config_user_id;

-- Drop the table
DROP TABLE IF EXISTS app_config;

-- Note: The update_updated_at_column function might still be needed by other tables,
-- so it's typically dropped in the *very last* down migration if it's shared.
-- If this IS the last table using it, uncomment the line below.
-- DROP FUNCTION IF EXISTS update_updated_at_column(); 