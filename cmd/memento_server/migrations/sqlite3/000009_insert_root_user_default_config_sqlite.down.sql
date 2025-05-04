-- Delete default config for root user
DELETE FROM app_config
WHERE user_id = (SELECT x'00000000000000000000000000000001');
