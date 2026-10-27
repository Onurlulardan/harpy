-- +goose Up
CREATE TABLE iam_roles (
    id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    organization_id uuid,
    name text NOT NULL,
    name_key bytea NOT NULL,
    description text,
    status text NOT NULL,
    is_protected boolean NOT NULL,
    is_default boolean NOT NULL,
    profile_resource_kind text NOT NULL,
    profile_schema_version_id uuid NOT NULL,
    profile_values jsonb NOT NULL,
    profile_sha256 bytea NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    retired_at timestamptz,
    CONSTRAINT iam_roles_pkey PRIMARY KEY (id),
    CONSTRAINT iam_roles_scope_identity_key UNIQUE (id, scope_kind, scope_id),
    CONSTRAINT iam_roles_organization_fk FOREIGN KEY (organization_id)
        REFERENCES iam_organizations (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_roles_profile_schema_fk
        FOREIGN KEY (profile_resource_kind, profile_schema_version_id)
        REFERENCES iam_profile_schema_versions (resource_kind, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_roles_scope_check CHECK (
        (scope_kind = 'tenant' AND organization_id IS NULL) OR
        (scope_kind = 'organization' AND organization_id IS NOT NULL AND
            organization_id = scope_id)
    ),
    CONSTRAINT iam_roles_name_check CHECK (
        pg_catalog.char_length(name) BETWEEN 1 AND 255 AND name = pg_catalog.btrim(name)
    ),
    CONSTRAINT iam_roles_name_key_check CHECK (
        pg_catalog.octet_length(name_key) BETWEEN 1 AND 1024 AND
        name_key = pg_catalog.convert_to(name, 'UTF8')
    ),
    CONSTRAINT iam_roles_description_check CHECK (
        description IS NULL OR pg_catalog.char_length(description) <= 2000
    ),
    CONSTRAINT iam_roles_status_check CHECK (status IN ('active', 'retired')),
    CONSTRAINT iam_roles_profile_resource_kind_check
        CHECK (profile_resource_kind = 'role'),
    CONSTRAINT iam_roles_profile_values_check
        CHECK (pg_catalog.jsonb_typeof(profile_values) = 'object'),
    CONSTRAINT iam_roles_profile_sha256_check
        CHECK (pg_catalog.octet_length(profile_sha256) = 32),
    CONSTRAINT iam_roles_revision_check CHECK (revision > 0),
    CONSTRAINT iam_roles_timestamp_order_check CHECK (updated_at >= created_at),
    CONSTRAINT iam_roles_retirement_check CHECK (
        (status = 'active' AND retired_at IS NULL) OR
        (status = 'retired' AND retired_at IS NOT NULL AND retired_at >= updated_at)
    )
);

CREATE UNIQUE INDEX iam_roles_tenant_name_uidx
    ON iam_roles (name_key) WHERE scope_kind = 'tenant';

CREATE UNIQUE INDEX iam_roles_organization_name_uidx
    ON iam_roles (organization_id, name_key)
    WHERE scope_kind = 'organization';

CREATE UNIQUE INDEX iam_roles_one_tenant_default_uidx
    ON iam_roles ((1))
    WHERE scope_kind = 'tenant' AND status = 'active' AND is_default;

CREATE UNIQUE INDEX iam_roles_one_organization_default_uidx
    ON iam_roles (organization_id)
    WHERE scope_kind = 'organization' AND status = 'active' AND is_default;
