-- +goose Up
CREATE TABLE iam_lifecycle_operation_receipts (
    operation_id uuid NOT NULL,
    command_key text COLLATE "C" NOT NULL,
    request_sha256 bytea NOT NULL,
    result_id uuid NOT NULL,
    subject_user_id uuid NOT NULL,
    result_revision bigint NOT NULL,
    token_version bigint NOT NULL,
    authorization_version bigint NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_lifecycle_operation_receipts_pkey PRIMARY KEY (operation_id),
    CONSTRAINT iam_lifecycle_operation_receipts_subject_fk
        FOREIGN KEY (subject_user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_lifecycle_operation_receipts_command_check CHECK (
        command_key ~ '^iam\.[a-z0-9_.]+$' AND
        pg_catalog.char_length(command_key) BETWEEN 5 AND 128
    ),
    CONSTRAINT iam_lifecycle_operation_receipts_hash_check
        CHECK (pg_catalog.octet_length(request_sha256) = 32),
    CONSTRAINT iam_lifecycle_operation_receipts_versions_check CHECK (
        result_revision > 0 AND token_version > 0 AND authorization_version >= 0
    )
);

CREATE TABLE iam_lifecycle_audit (
    operation_id uuid NOT NULL,
    command_key text COLLATE "C" NOT NULL,
    source_kind text NOT NULL,
    actor_id uuid NOT NULL,
    subject_user_id uuid NOT NULL,
    resource_id uuid NOT NULL,
    resource_kind text COLLATE "C" NOT NULL,
    resource_revision bigint NOT NULL,
    token_version bigint NOT NULL,
    authorization_version bigint NOT NULL,
    request_sha256 bytea NOT NULL,
    occurred_at timestamptz NOT NULL,
    CONSTRAINT iam_lifecycle_audit_pkey PRIMARY KEY (operation_id),
    CONSTRAINT iam_lifecycle_audit_receipt_fk FOREIGN KEY (operation_id)
        REFERENCES iam_lifecycle_operation_receipts (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_lifecycle_audit_actor_fk FOREIGN KEY (actor_id)
        REFERENCES iam_users (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_lifecycle_audit_subject_fk FOREIGN KEY (subject_user_id)
        REFERENCES iam_users (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_lifecycle_audit_source_check
        CHECK (source_kind IN ('actor_tenant', 'credential_target')),
    CONSTRAINT iam_lifecycle_audit_kind_check CHECK (
        resource_kind ~ '^[a-z][a-z0-9_.]{1,63}$'
    ),
    CONSTRAINT iam_lifecycle_audit_hash_check
        CHECK (pg_catalog.octet_length(request_sha256) = 32),
    CONSTRAINT iam_lifecycle_audit_versions_check CHECK (
        resource_revision > 0 AND token_version > 0 AND
        authorization_version >= 0
    )
);

CREATE INDEX iam_lifecycle_audit_subject_idx
    ON iam_lifecycle_audit (subject_user_id, occurred_at DESC);
