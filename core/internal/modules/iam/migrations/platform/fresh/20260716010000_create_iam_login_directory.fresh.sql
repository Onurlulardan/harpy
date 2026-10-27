-- +goose Up
CREATE TABLE platform.iam_login_directory (
    tenant_id uuid NOT NULL,
    identifier_kind text NOT NULL,
    lookup_key_id text COLLATE "C" NOT NULL,
    identifier_hmac bytea NOT NULL,
    user_id uuid NOT NULL,
    source_identifier_id uuid NOT NULL,
    source_revision bigint NOT NULL,
    state text NOT NULL,
    source_sha256 bytea NOT NULL,
    projected_at timestamptz NOT NULL,
    tombstoned_at timestamptz,
    CONSTRAINT iam_login_directory_pkey PRIMARY KEY (
        tenant_id, identifier_kind, lookup_key_id, identifier_hmac
    ),
    CONSTRAINT iam_login_directory_source_unique UNIQUE (
        tenant_id, lookup_key_id, source_identifier_id
    ),
    CONSTRAINT iam_login_directory_tenant_fk FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_login_directory_identifier_kind_check CHECK (
        identifier_kind IN ('email_v1', 'portal_v1')
    ),
    CONSTRAINT iam_login_directory_lookup_key_check CHECK (
        pg_catalog.octet_length(lookup_key_id) BETWEEN 1 AND 128 AND
        lookup_key_id ~ '^[a-z0-9][a-z0-9._-]{0,127}$'
    ),
    CONSTRAINT iam_login_directory_identifier_hmac_check CHECK (
        pg_catalog.octet_length(identifier_hmac) = 32
    ),
    CONSTRAINT iam_login_directory_source_revision_check CHECK (
        source_revision > 0
    ),
    CONSTRAINT iam_login_directory_state_check CHECK (
        state IN ('active', 'tombstoned')
    ),
    CONSTRAINT iam_login_directory_source_sha256_check CHECK (
        pg_catalog.octet_length(source_sha256) = 32
    ),
    CONSTRAINT iam_login_directory_tombstone_check CHECK (
        (state = 'active' AND tombstoned_at IS NULL) OR
        (state = 'tombstoned' AND tombstoned_at IS NOT NULL AND
            projected_at <= tombstoned_at)
    )
);

CREATE INDEX iam_login_directory_user_idx
    ON platform.iam_login_directory (tenant_id, user_id, identifier_kind, state);

CREATE INDEX iam_login_directory_tombstone_idx
    ON platform.iam_login_directory (tenant_id, tombstoned_at)
    WHERE state = 'tombstoned';
