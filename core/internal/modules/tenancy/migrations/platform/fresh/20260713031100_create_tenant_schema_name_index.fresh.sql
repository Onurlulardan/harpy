-- +goose Up
CREATE UNIQUE INDEX tenancy_tenants_schema_name_key
ON public.tenants (schema_name);
