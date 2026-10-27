-- +goose Up
CREATE TABLE tenancy_schema_identity (
    singleton boolean PRIMARY KEY CHECK (singleton),
    tenant_id uuid NOT NULL UNIQUE,
    schema_name text NOT NULL,
    origin text NOT NULL CHECK (origin IN ('new', 'v1_adopted')),
    created_at timestamptz NOT NULL
);
