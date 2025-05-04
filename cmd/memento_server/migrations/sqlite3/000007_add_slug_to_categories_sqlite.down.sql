    -- Drop the unique index
    -- Note: Dropping the column itself requires complex table recreation
    DROP INDEX IF EXISTS uk_categories_user_slug;