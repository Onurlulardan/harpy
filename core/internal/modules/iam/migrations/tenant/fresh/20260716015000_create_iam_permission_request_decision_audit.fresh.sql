-- +goose Up
CREATE TABLE iam_permission_request_decision_audit (
    operation_id uuid NOT NULL,
    group_id uuid NOT NULL,
    reviewer_user_id uuid NOT NULL,
    review_task_id uuid NOT NULL,
    decision text COLLATE "C" NOT NULL,
    result_code text COLLATE "C" NOT NULL,
    request_sha256 bytea NOT NULL,
    group_revision bigint NOT NULL,
    authorization_clock bigint NOT NULL,
    recorded_at timestamptz NOT NULL,
    CONSTRAINT iam_permission_request_decision_audit_pkey
        PRIMARY KEY (operation_id),
    CONSTRAINT iam_permission_request_decision_audit_receipt_fk
        FOREIGN KEY (operation_id)
        REFERENCES iam_authorization_operation_receipts (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_request_decision_audit_group_fk
        FOREIGN KEY (group_id) REFERENCES iam_permission_request_groups (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_request_decision_audit_reviewer_fk
        FOREIGN KEY (reviewer_user_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_permission_request_decision_audit_decision_check
        CHECK (decision IN ('APPROVE', 'REJECT')),
    CONSTRAINT iam_permission_request_decision_audit_result_check CHECK (
        (decision = 'APPROVE' AND
            result_code = 'iam.permission_request.approved') OR
        (decision = 'REJECT' AND
            result_code = 'iam.permission_request.rejected')
    ),
    CONSTRAINT iam_permission_request_decision_audit_hash_check
        CHECK (pg_catalog.octet_length(request_sha256) = 32),
    CONSTRAINT iam_permission_request_decision_audit_version_check
        CHECK (group_revision > 0 AND authorization_clock > 0)
);

CREATE INDEX iam_permission_request_decision_audit_group_idx
    ON iam_permission_request_decision_audit (group_id, recorded_at DESC);

CREATE TRIGGER iam_permission_request_decision_audit_mutation_guard
BEFORE UPDATE OR DELETE OR TRUNCATE
ON iam_permission_request_decision_audit
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_permission_request_decision_audit ENABLE ALWAYS TRIGGER
    iam_permission_request_decision_audit_mutation_guard;
