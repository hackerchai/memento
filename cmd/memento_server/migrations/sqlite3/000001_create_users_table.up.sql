-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id BLOB PRIMARY KEY DEFAULT (randomblob(16)), -- Store UUID as BLOB
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role INTEGER NOT NULL DEFAULT 0, -- 0: User, 1: Admin, 2: Root
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f', 'now')),
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%d %H:%M:%f', 'now')),
    totp_secret TEXT NULL,
    passkey_data BLOB NULL,
    third_party_provider TEXT NULL,
    third_party_user_id TEXT NULL
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Automatically update updated_at timestamp using a trigger
CREATE TRIGGER IF NOT EXISTS update_users_updated_at
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
    UPDATE users SET updated_at = strftime('%Y-%m-%d %H:%M:%f', 'now') WHERE id = OLD.id;
END;

-- Insert root user (password: "root")
-- Hash generated using argon2id (m=65536, t=3, p=4) with salt "aaaaaaaaaaaaaaaaaaaaaa"
-- Note: We use hex representation for the fixed root user ID BLOB.
INSERT OR IGNORE INTO users (id, name, email, password, role)
VALUES (
    X'00000000000000000000000000000001', -- Fixed UUID for root user (hex blob)
    'root',
    'root@localhost.local',
    '$argon2id$v=19$m=65536,t=3,p=4$YWFhYWFhYWFhYWFhYWFhYQ$Fq6v1NlqJ/0P5p3Xh8z7kM3j0iW4T6YtQv3o1p5L9bA',
    2 -- RoleRoot
);
