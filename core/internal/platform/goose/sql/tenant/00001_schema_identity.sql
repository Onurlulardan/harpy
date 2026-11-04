-- +goose Up
CREATE TABLE tenancy_schema_identity (
    singleton boolean PRIMARY KEY,
    tenant_id uuid NOT NULL,
    schema_name text NOT NULL,
    origin text NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT tenancy_schema_identity_singleton CHECK (singleton),
    CONSTRAINT tenancy_schema_identity_origin CHECK (origin IN ('new', 'v1_adopted')),
    CONSTRAINT tenancy_schema_identity_schema_name CHECK (
        schema_name ~ '^tenant_[0-9a-f]{32}$'
    )
);

-- +goose Down
DROP TABLE tenancy_schema_identity;
