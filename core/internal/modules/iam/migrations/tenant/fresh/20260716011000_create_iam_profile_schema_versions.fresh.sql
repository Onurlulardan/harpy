-- +goose Up
CREATE TABLE iam_profile_schema_versions (
    id uuid NOT NULL,
    resource_kind text NOT NULL,
    version bigint NOT NULL,
    state text NOT NULL,
    definition_sha256 bytea NOT NULL,
    activated_at timestamptz,
    retired_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    revision bigint NOT NULL,
    CONSTRAINT iam_profile_schema_versions_pkey PRIMARY KEY (id),
    CONSTRAINT iam_profile_schema_versions_resource_version_key
        UNIQUE (resource_kind, version),
    CONSTRAINT iam_profile_schema_versions_resource_id_key
        UNIQUE (resource_kind, id),
    CONSTRAINT iam_profile_schema_versions_resource_kind_check
        CHECK (resource_kind IN ('user', 'organization', 'role')),
    CONSTRAINT iam_profile_schema_versions_version_check CHECK (version > 0),
    CONSTRAINT iam_profile_schema_versions_state_check
        CHECK (state IN ('draft', 'active', 'retired')),
    CONSTRAINT iam_profile_schema_versions_definition_sha256_check
        CHECK (pg_catalog.octet_length(definition_sha256) = 32),
    CONSTRAINT iam_profile_schema_versions_revision_check CHECK (revision > 0),
    CONSTRAINT iam_profile_schema_versions_timestamp_order_check CHECK (
        updated_at >= created_at AND
        (activated_at IS NULL OR activated_at >= created_at) AND
        (retired_at IS NULL OR retired_at >= updated_at)
    ),
    CONSTRAINT iam_profile_schema_versions_state_timestamps_check CHECK (
        (state = 'draft' AND activated_at IS NULL AND retired_at IS NULL)
        OR (state = 'active' AND activated_at IS NOT NULL AND retired_at IS NULL)
        OR (
            state = 'retired'
            AND activated_at IS NOT NULL
            AND retired_at IS NOT NULL
            AND retired_at >= activated_at
        )
    )
);

CREATE UNIQUE INDEX iam_profile_schema_versions_one_active_resource_idx
    ON iam_profile_schema_versions (resource_kind)
    WHERE state = 'active';
