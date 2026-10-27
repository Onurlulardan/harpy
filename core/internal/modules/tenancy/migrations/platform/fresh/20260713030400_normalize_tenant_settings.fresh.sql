-- +goose Up
UPDATE public.tenants
SET settings = '{}'::jsonb
WHERE settings IS NULL;
