-- Delete default config for root user
DELETE FROM app_config
WHERE user_id = '00000000-0000-0000-0000-000000000001';
