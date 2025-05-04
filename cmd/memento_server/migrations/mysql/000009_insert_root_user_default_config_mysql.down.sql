-- Delete default config for root user
DELETE FROM app_config
WHERE user_id = UUID_TO_BIN('00000000-0000-0000-0000-000000000001');
