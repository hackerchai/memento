-- +migrate Down
ALTER TABLE articles
DROP COLUMN is_read,
DROP COLUMN is_starred,
DROP COLUMN original_html;
