-- +goose Up
CREATE TABLE iam_organizations (
    id uuid NOT NULL,
    name text NOT NULL,
    slug text COLLATE "C" NOT NULL,
    status text NOT NULL,
    parent_id uuid,
    profile_resource_kind text NOT NULL,
    profile_schema_version_id uuid NOT NULL,
    profile_values jsonb NOT NULL,
    profile_sha256 bytea NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    tombstoned_at timestamptz,
    CONSTRAINT iam_organizations_pkey PRIMARY KEY (id),
    CONSTRAINT iam_organizations_slug_key UNIQUE (slug),
    CONSTRAINT iam_organizations_parent_fk FOREIGN KEY (parent_id)
        REFERENCES iam_organizations (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_organizations_profile_schema_fk
        FOREIGN KEY (profile_resource_kind, profile_schema_version_id)
        REFERENCES iam_profile_schema_versions (resource_kind, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_organizations_name_check CHECK (
        pg_catalog.char_length(name) BETWEEN 1 AND 255 AND name = pg_catalog.btrim(name)
    ),
    CONSTRAINT iam_organizations_slug_check CHECK (
        pg_catalog.char_length(slug) BETWEEN 1 AND 100 AND
        slug = pg_catalog.lower(slug) AND
        (slug ~ '^[a-z0-9]$' OR slug ~ '^[a-z0-9][a-z0-9-]*[a-z0-9]$')
    ),
    CONSTRAINT iam_organizations_status_check
        CHECK (status IN ('active', 'suspended', 'tombstoned')),
    CONSTRAINT iam_organizations_parent_check CHECK (parent_id IS DISTINCT FROM id),
    CONSTRAINT iam_organizations_profile_resource_kind_check
        CHECK (profile_resource_kind = 'organization'),
    CONSTRAINT iam_organizations_profile_values_check
        CHECK (pg_catalog.jsonb_typeof(profile_values) = 'object'),
    CONSTRAINT iam_organizations_profile_sha256_check
        CHECK (pg_catalog.octet_length(profile_sha256) = 32),
    CONSTRAINT iam_organizations_revision_check CHECK (revision > 0),
    CONSTRAINT iam_organizations_timestamp_order_check CHECK (updated_at >= created_at),
    CONSTRAINT iam_organizations_tombstone_check CHECK (
        (status <> 'tombstoned' AND tombstoned_at IS NULL) OR
        (status = 'tombstoned' AND tombstoned_at IS NOT NULL AND
            tombstoned_at >= updated_at)
    )
);

CREATE INDEX iam_organizations_parent_idx ON iam_organizations (parent_id);
