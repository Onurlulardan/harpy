-- +goose Up
ALTER TABLE public.tenants
    ALTER COLUMN created_at TYPE timestamptz
        USING created_at AT TIME ZONE 'UTC',
    ALTER COLUMN updated_at TYPE timestamptz
        USING updated_at AT TIME ZONE 'UTC',
    ALTER COLUMN deleted_at TYPE timestamptz
        USING deleted_at AT TIME ZONE 'UTC';
