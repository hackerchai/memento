-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id BINARY(16) PRIMARY KEY DEFAULT (UUID_TO_BIN(UUID())),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role INT NOT NULL DEFAULT 0, -- 0: User, 1: Admin, 2: Root
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    totp_secret VARCHAR(255) NULL,
    passkey_data BLOB NULL,
    third_party_provider VARCHAR(50) NULL,
    third_party_user_id VARCHAR(255) NULL,
    INDEX idx_users_email (email)
);

-- Insert root user (password: "root")
-- Hash generated using argon2id (m=65536, t=3, p=4) with salt "aaaaaaaaaaaaaaaaaaaaaa"
-- Note: UUID_TO_BIN('00000000-0000-0000-0000-000000000001') is used for the fixed root user ID.
INSERT IGNORE INTO users (id, name, email, password, role)
VALUES (
    UUID_TO_BIN('00000000-0000-0000-0000-000000000001'),
    'root',
    'root@localhost.local',
    '$argon2id$v=19$m=65536,t=3,p=4$YWFhYWFhYWFhYWFhYWFhYQ$Fq6v1NlqJ/0P5p3Xh8z7kM3j0iW4T6YtQv3o1p5L9bA',
    2 -- RoleRoot
);
