-- +goose Up
CREATE TABLE platform.tenants (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    slug text NOT NULL,
    status text NOT NULL,
    schema_name text NOT NULL,
    settings_revision bigint NOT NULL,
    lifecycle_revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz,
    offboarded_at timestamptz,
    purged_at timestamptz,
    CONSTRAINT tenants_name_length CHECK (char_length(name) BETWEEN 1 AND 255),
    CONSTRAINT tenants_slug_shape CHECK (
        slug ~ '^[a-z0-9](?:[a-z0-9-]{0,98}[a-z0-9])?$'
    ),
    CONSTRAINT tenants_status_check CHECK (
        status IN ('PROVISIONING', 'ACTIVE', 'TRIAL', 'SUSPENDED')
    ),
    CONSTRAINT tenants_schema_name_check CHECK (schema_name ~ '^t_[a-z2-7]{26}$'),
    CONSTRAINT tenants_settings_revision_check CHECK (settings_revision > 0),
    CONSTRAINT tenants_lifecycle_revision_check CHECK (lifecycle_revision > 0)
);

CREATE UNIQUE INDEX tenants_schema_name_key ON platform.tenants (schema_name);

CREATE UNIQUE INDEX tenants_slug_key ON platform.tenants (slug);

-- +goose Down
DROP TABLE platform.tenants;
