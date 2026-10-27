-- +goose Up
CREATE TABLE iam_workflow_permission_decision_receipts (
    operation_id uuid NOT NULL,
    effect_key text COLLATE "C" NOT NULL,
    request_sha256 bytea NOT NULL,
    request_id uuid NOT NULL,
    request_revision_before bigint NOT NULL,
    request_revision_after bigint NOT NULL,
    group_id uuid NOT NULL,
    group_revision_before bigint NOT NULL,
    group_revision_after bigint NOT NULL,
    decision text NOT NULL,
    principal_id uuid NOT NULL,
    principal_revision bigint NOT NULL,
    principal_authorization_version bigint NOT NULL,
    grant_id uuid NOT NULL,
    grant_revision bigint NOT NULL,
    definition_id uuid NOT NULL,
    manifest_sha256 bytea NOT NULL,
    plan_sha256 bytea NOT NULL,
    verified_at timestamptz NOT NULL,
    decided_at timestamptz NOT NULL,
    authorization_clock bigint NOT NULL,
    result_code text COLLATE "C" NOT NULL,
    receipt_document jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_workflow_permission_decision_receipts_pkey
        PRIMARY KEY (operation_id),
    CONSTRAINT iam_workflow_permission_decision_receipts_effect_key
        UNIQUE (effect_key),
    CONSTRAINT iam_workflow_permission_decision_receipts_operation_fk
        FOREIGN KEY (operation_id)
        REFERENCES iam_authorization_operation_receipts (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_workflow_permission_decision_receipts_request_fk
        FOREIGN KEY (request_id) REFERENCES iam_permission_requests (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_workflow_permission_decision_receipts_group_fk
        FOREIGN KEY (group_id) REFERENCES iam_permission_request_groups (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_workflow_permission_decision_receipts_principal_fk
        FOREIGN KEY (principal_id) REFERENCES iam_service_principals (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_workflow_permission_decision_receipts_grant_fk
        FOREIGN KEY (grant_id, grant_revision)
        REFERENCES iam_capability_grant_revisions (grant_id, grant_revision)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_workflow_permission_decision_receipts_effect_check CHECK (
        pg_catalog.char_length(effect_key) BETWEEN 1 AND 256 AND
        effect_key ~ '^[A-Za-z0-9][A-Za-z0-9_.:/-]*$'
    ),
    CONSTRAINT iam_workflow_permission_decision_receipts_hash_check CHECK (
        pg_catalog.octet_length(request_sha256) = 32 AND
        pg_catalog.octet_length(manifest_sha256) = 32 AND
        pg_catalog.octet_length(plan_sha256) = 32
    ),
    CONSTRAINT iam_workflow_permission_decision_receipts_revision_check CHECK (
        request_revision_before > 0 AND
        request_revision_after = request_revision_before + 1 AND
        group_revision_before > 0 AND
        group_revision_after = group_revision_before + 1 AND
        principal_revision > 0 AND principal_authorization_version > 0 AND
        grant_revision > 0 AND authorization_clock > 0
    ),
    CONSTRAINT iam_workflow_permission_decision_receipts_decision_check
        CHECK (decision IN ('APPROVE', 'REJECT')),
    CONSTRAINT iam_workflow_permission_decision_receipts_result_check CHECK (
        result_code IN (
            'iam.permission_request.approved',
            'iam.permission_request.rejected'
        )
    ),
    CONSTRAINT iam_workflow_permission_decision_receipts_time_check CHECK (
        decided_at >= verified_at AND created_at = decided_at
    ),
    CONSTRAINT iam_workflow_permission_decision_receipts_document_check CHECK (
        pg_catalog.jsonb_typeof(receipt_document) = 'object' AND
        pg_catalog.octet_length(receipt_document::text) BETWEEN 2 AND 65536
    )
);

CREATE TRIGGER iam_workflow_permission_decision_receipts_mutation_guard
BEFORE UPDATE OR DELETE OR TRUNCATE
ON iam_workflow_permission_decision_receipts
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();

ALTER TABLE iam_workflow_permission_decision_receipts ENABLE ALWAYS TRIGGER
    iam_workflow_permission_decision_receipts_mutation_guard;
