-- +goose Up
CREATE TABLE iam_rbac_change_audit (
    operation_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    aggregate_id uuid NOT NULL,
    request_sha256 bytea NOT NULL,
    command_key text COLLATE "C" NOT NULL,
    result_code text COLLATE "C" NOT NULL,
    aggregate_revision bigint NOT NULL,
    authorization_clock bigint NOT NULL,
    recorded_at timestamptz NOT NULL,
    CONSTRAINT iam_rbac_change_audit_pkey PRIMARY KEY (operation_id),
    CONSTRAINT iam_rbac_change_audit_receipt_fk FOREIGN KEY (operation_id)
        REFERENCES iam_authorization_operation_receipts (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_rbac_change_audit_actor_fk FOREIGN KEY (actor_id)
        REFERENCES iam_users (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_rbac_change_audit_request_check CHECK (
        pg_catalog.octet_length(request_sha256) = 32
    ),
    CONSTRAINT iam_rbac_change_audit_command_check CHECK (
        command_key ~ '^iam\.rbac\.[a-z0-9_.]+$'
    ),
    CONSTRAINT iam_rbac_change_audit_result_check CHECK (
        result_code ~ '^iam\.rbac\.[a-z0-9_.]+$'
    ),
    CONSTRAINT iam_rbac_change_audit_revision_check CHECK (aggregate_revision > 0),
    CONSTRAINT iam_rbac_change_audit_clock_check CHECK (authorization_clock > 0)
);

CREATE TRIGGER iam_rbac_change_audit_mutation_guard
BEFORE UPDATE OR DELETE OR TRUNCATE ON iam_rbac_change_audit
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_rbac_change_audit
    ENABLE ALWAYS TRIGGER iam_rbac_change_audit_mutation_guard;
