-- +goose Up
CREATE TABLE iam_password_credentials (
    user_id uuid NOT NULL,
    credential_version bigint NOT NULL,
    state text NOT NULL,
    algorithm text NOT NULL,
    salt bytea,
    tag bytea,
    memory_kib integer,
    iterations smallint,
    parallelism smallint,
    legacy_bcrypt_hash text,
    upgrade_required boolean NOT NULL,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    last_used_at timestamptz,
    invalidated_at timestamptz,
    CONSTRAINT iam_password_credentials_pkey PRIMARY KEY (user_id),
    CONSTRAINT iam_password_credentials_user_fk FOREIGN KEY (user_id)
        REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_password_credentials_version_check
        CHECK (credential_version > 0),
    CONSTRAINT iam_password_credentials_state_check
        CHECK (state IN ('active', 'reset_required', 'disabled', 'compromised')),
    CONSTRAINT iam_password_credentials_algorithm_check
        CHECK (algorithm IN ('argon2id_v1', 'bcrypt_v1_adopted')),
    CONSTRAINT iam_password_credentials_shape_check CHECK (
        (algorithm = 'argon2id_v1' AND
            salt IS NOT NULL AND tag IS NOT NULL AND memory_kib IS NOT NULL AND
            iterations IS NOT NULL AND parallelism IS NOT NULL AND
            pg_catalog.octet_length(salt) = 16 AND
            pg_catalog.octet_length(tag) = 32 AND
            memory_kib = 65536 AND iterations = 3 AND parallelism = 4 AND
            legacy_bcrypt_hash IS NULL AND NOT upgrade_required) OR
        (algorithm = 'bcrypt_v1_adopted' AND
            salt IS NULL AND tag IS NULL AND memory_kib IS NULL AND
            iterations IS NULL AND parallelism IS NULL AND
            legacy_bcrypt_hash IS NOT NULL AND
            pg_catalog.char_length(legacy_bcrypt_hash) = 60 AND
            legacy_bcrypt_hash ~ '^\$2[aby]\$(0[4-9]|[12][0-9]|3[01])\$[./A-Za-z0-9]{53}$' AND
            upgrade_required)
    ),
    CONSTRAINT iam_password_credentials_revision_check CHECK (revision > 0),
    CONSTRAINT iam_password_credentials_timestamp_order_check CHECK (
        updated_at >= created_at AND
        (last_used_at IS NULL OR last_used_at >= created_at) AND
        (invalidated_at IS NULL OR invalidated_at >= created_at)
    ),
    CONSTRAINT iam_password_credentials_invalidation_check CHECK (
        (state IN ('active', 'reset_required') AND invalidated_at IS NULL) OR
        (state IN ('disabled', 'compromised') AND invalidated_at IS NOT NULL)
    )
);
