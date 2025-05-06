-- +migrate Up
-- SQLite requires adding columns one by one
ALTER TABLE articles ADD COLUMN is_read BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE articles ADD COLUMN is_starred BOOLEAN NOT NULL DEFAULT 0;
ALTER TABLE articles ADD COLUMN original_html TEXT;
