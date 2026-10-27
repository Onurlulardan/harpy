-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_slug
ON public.tenants (slug);
