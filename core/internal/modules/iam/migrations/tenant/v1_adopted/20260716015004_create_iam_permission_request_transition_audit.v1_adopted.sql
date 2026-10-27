-- +goose Up
CREATE TABLE iam_permission_request_transition_receipts (
    operation_id uuid NOT NULL,
    command_key text COLLATE "C" NOT NULL,
    request_sha256 bytea NOT NULL,
    result_code text COLLATE "C" NOT NULL,
    result_id uuid NOT NULL,
    result_revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_permission_request_transition_receipts_pkey
        PRIMARY KEY (operation_id),
    CONSTRAINT iam_permission_request_transition_receipts_hash_check
        CHECK (pg_catalog.octet_length(request_sha256) = 32),
    CONSTRAINT iam_permission_request_transition_receipts_revision_check
        CHECK (result_revision > 0)
);

CREATE TABLE iam_permission_request_transition_audit (
    operation_id uuid NOT NULL,
    group_id uuid NOT NULL,
    actor_user_id uuid NOT NULL,
    review_task_id uuid,
    command_key text COLLATE "C" NOT NULL,
    from_state text COLLATE "C",
    to_state text COLLATE "C" NOT NULL,
    request_sha256 bytea NOT NULL,
    group_revision bigint NOT NULL,
    authorization_clock bigint NOT NULL,
    recorded_at timestamptz NOT NULL,
    CONSTRAINT iam_permission_request_transition_audit_pkey
        PRIMARY KEY (operation_id),
    CONSTRAINT iam_permission_request_transition_audit_receipt_fk
        FOREIGN KEY (operation_id)
        REFERENCES iam_permission_request_transition_receipts (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_request_transition_audit_group_fk
        FOREIGN KEY (group_id) REFERENCES iam_permission_request_groups (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_request_transition_audit_actor_fk
        FOREIGN KEY (actor_user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_request_transition_audit_hash_check
        CHECK (pg_catalog.octet_length(request_sha256) = 32),
    CONSTRAINT iam_permission_request_transition_audit_version_check
        CHECK (group_revision > 0 AND authorization_clock >= 0),
    CONSTRAINT iam_permission_request_transition_audit_shape_check CHECK (
        (command_key = 'iam.permission_request.create' AND
            from_state IS NULL AND to_state = 'PENDING' AND
            review_task_id IS NULL AND group_revision = 1) OR
        (command_key = 'iam.permission_request.review' AND
            from_state = 'PENDING' AND to_state = 'IN_REVIEW' AND
            review_task_id IS NOT NULL) OR
        (command_key = 'iam.permission_request.cancel' AND
            from_state IN ('PENDING', 'IN_REVIEW') AND
            to_state = 'CANCELLED' AND review_task_id IS NULL) OR
        (command_key = 'iam.permission_request.expire' AND
            from_state IN ('PENDING', 'IN_REVIEW') AND
            to_state = 'EXPIRED' AND review_task_id IS NULL)
    )
);

CREATE INDEX iam_permission_request_transition_audit_group_idx
    ON iam_permission_request_transition_audit (group_id, recorded_at DESC);

CREATE TRIGGER iam_permission_request_transition_receipt_mutation_guard
BEFORE UPDATE OR DELETE OR TRUNCATE
ON iam_permission_request_transition_receipts
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_permission_request_transition_receipts ENABLE ALWAYS TRIGGER
    iam_permission_request_transition_receipt_mutation_guard;

CREATE TRIGGER iam_permission_request_transition_audit_mutation_guard
BEFORE UPDATE OR DELETE OR TRUNCATE
ON iam_permission_request_transition_audit
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_permission_request_transition_audit ENABLE ALWAYS TRIGGER
    iam_permission_request_transition_audit_mutation_guard;
