-- +goose Up
CREATE VIEW public.tenants AS
SELECT
    id,
    name,
    slug,
    status,
    settings,
    owner_id,
    created_at,
    updated_at,
    deleted_at
FROM platform.tenants;
