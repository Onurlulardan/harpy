-- +goose Up
ALTER TABLE public.tenants
    ALTER COLUMN schema_name SET NOT NULL,
    ADD CONSTRAINT tenancy_tenants_status_check
        CHECK (status IN ('ACTIVE', 'TRIAL', 'SUSPENDED')),
    ADD CONSTRAINT tenancy_tenants_schema_name_check CHECK (
        pg_catalog.octet_length(schema_name) BETWEEN 1 AND 63
        AND schema_name ~ '^[a-z][a-z0-9_]*$'
    );
