-- +goose Up
ALTER TABLE public.tenants
    ALTER COLUMN settings SET DEFAULT '{}'::jsonb,
    ALTER COLUMN settings SET NOT NULL;
