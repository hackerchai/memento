-- Remove locale column from app_config table
ALTER TABLE app_config
DROP COLUMN locale;
