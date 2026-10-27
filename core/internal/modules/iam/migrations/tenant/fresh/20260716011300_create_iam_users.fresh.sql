-- +goose Up
CREATE TABLE iam_users (
    id uuid NOT NULL,
    account_type text NOT NULL,
    status text NOT NULL,
    profile_resource_kind text NOT NULL,
    profile_schema_version_id uuid NOT NULL,
    profile_values jsonb NOT NULL,
    profile_sha256 bytea NOT NULL,
    token_version bigint NOT NULL,
    authorization_version bigint NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    tombstoned_at timestamptz,
    CONSTRAINT iam_users_pkey PRIMARY KEY (id),
    CONSTRAINT iam_users_profile_schema_fk
        FOREIGN KEY (profile_resource_kind, profile_schema_version_id)
        REFERENCES iam_profile_schema_versions (resource_kind, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_users_account_type_check
        CHECK (account_type IN ('operator', 'portal')),
    CONSTRAINT iam_users_status_check
        CHECK (status IN ('pending', 'active', 'suspended', 'disabled', 'tombstoned')),
    CONSTRAINT iam_users_profile_resource_kind_check
        CHECK (profile_resource_kind = 'user'),
    CONSTRAINT iam_users_profile_values_check
        CHECK (pg_catalog.jsonb_typeof(profile_values) = 'object'),
    CONSTRAINT iam_users_profile_sha256_check
        CHECK (pg_catalog.octet_length(profile_sha256) = 32),
    CONSTRAINT iam_users_token_version_check CHECK (token_version > 0),
    CONSTRAINT iam_users_authorization_version_check
        CHECK (authorization_version >= 0),
    CONSTRAINT iam_users_revision_check CHECK (revision > 0),
    CONSTRAINT iam_users_timestamp_order_check CHECK (updated_at >= created_at),
    CONSTRAINT iam_users_tombstone_check CHECK (
        (status <> 'tombstoned' AND tombstoned_at IS NULL) OR
        (status = 'tombstoned' AND tombstoned_at IS NOT NULL AND
            tombstoned_at >= updated_at)
    )
);

CREATE INDEX iam_users_status_idx ON iam_users (status, account_type);
