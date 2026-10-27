-- +goose Up
CREATE TABLE platform.environment_identity (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    environment_id uuid NOT NULL UNIQUE,
    database_origin text NOT NULL CHECK (database_origin IN ('fresh', 'v1_adopted')),
    database_system_identifier text NOT NULL
        CHECK (database_system_identifier ~ '^[1-9][0-9]{0,19}$'),
    database_oid bigint NOT NULL CHECK (database_oid BETWEEN 1 AND 4294967295),
    release_id uuid NOT NULL,
    bootstrap_authority_sha256 text NOT NULL
        CHECK (bootstrap_authority_sha256 ~ '^[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL DEFAULT statement_timestamp()
);
