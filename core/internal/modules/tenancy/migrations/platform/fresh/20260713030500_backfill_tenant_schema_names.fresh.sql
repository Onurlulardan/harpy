-- +goose Up
UPDATE public.tenants
SET schema_name = (('tenant_' || slug)::name)::text
WHERE schema_name IS NULL;
