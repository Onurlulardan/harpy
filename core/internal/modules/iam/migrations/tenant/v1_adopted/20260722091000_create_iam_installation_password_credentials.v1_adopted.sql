-- +goose Up
CREATE TABLE iam_installation_password_credentials (
    credential_id uuid NOT NULL,
    user_id uuid NOT NULL,
    installation_operation_id uuid NOT NULL,
    mutation_operation_id uuid NOT NULL,
    credential_version bigint NOT NULL,
    purpose text NOT NULL,
    state text NOT NULL,
    superseded_by uuid,
    revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    retired_at timestamptz,
    CONSTRAINT iam_installation_password_credentials_pkey
        PRIMARY KEY (credential_id),
    CONSTRAINT iam_installation_password_credentials_user_fk
        FOREIGN KEY (user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_installation_password_credentials_successor_fk
        FOREIGN KEY (superseded_by)
        REFERENCES iam_installation_password_credentials (credential_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_installation_password_credentials_user_version_unique
        UNIQUE (user_id, credential_version),
    CONSTRAINT iam_installation_password_credentials_mutation_unique
        UNIQUE (mutation_operation_id),
    CONSTRAINT iam_installation_password_credentials_version_check
        CHECK (credential_version OPERATOR(pg_catalog.>) 0),
    CONSTRAINT iam_installation_password_credentials_purpose_check
        CHECK (purpose OPERATOR(pg_catalog.=) 'installation_initial_v1'),
    CONSTRAINT iam_installation_password_credentials_state_check
        CHECK (state OPERATOR(pg_catalog.=) ANY (
            ARRAY['active'::text, 'superseded'::text, 'retired'::text]
        )),
    CONSTRAINT iam_installation_password_credentials_revision_check
        CHECK (revision OPERATOR(pg_catalog.>) 0),
    CONSTRAINT iam_installation_password_credentials_state_shape_check CHECK (
        (state OPERATOR(pg_catalog.=) 'active' AND
            superseded_by IS NULL AND retired_at IS NULL) OR
        (state OPERATOR(pg_catalog.=) 'superseded' AND
            superseded_by IS NOT NULL AND retired_at IS NOT NULL) OR
        (state OPERATOR(pg_catalog.=) 'retired' AND
            superseded_by IS NULL AND retired_at IS NOT NULL)
    ),
    CONSTRAINT iam_installation_password_credentials_time_check CHECK (
        updated_at OPERATOR(pg_catalog.>=) created_at AND
        (retired_at IS NULL OR retired_at OPERATOR(pg_catalog.>=) created_at)
    )
);

CREATE UNIQUE INDEX iam_installation_password_credentials_active_user_unique
    ON iam_installation_password_credentials (user_id)
    WHERE state OPERATOR(pg_catalog.=) 'active';

