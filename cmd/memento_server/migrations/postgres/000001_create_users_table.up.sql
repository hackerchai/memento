-- Enable UUID generation function if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role INT NOT NULL DEFAULT 0, -- 0: User, 1: Admin, 2: Root
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    totp_secret VARCHAR(255) NULL,
    passkey_data BYTEA NULL,
    third_party_provider VARCHAR(50) NULL,
    third_party_user_id VARCHAR(255) NULL
);

-- Add index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);

-- Automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Insert root user (password: "root")
-- Hash generated using argon2id (m=65536, t=3, p=4) with salt "aaaaaaaaaaaaaaaaaaaaaa"
INSERT INTO users (id, name, email, password, role)
VALUES (
    '00000000-0000-0000-0000-000000000001', -- Fixed UUID for root user
    'root',
    'root@localhost.local',
    '$argon2id$v=19$m=16,t=2,p=1$MTIzNDU2Nzg$sZdZ0NAeciPMrU1YVkD/rQ',
    2 -- RoleRoot
)
ON CONFLICT (email) DO NOTHING; -- Avoid errors if already exists
