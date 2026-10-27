-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_forbid_identity_change_mutation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '42501',
        MESSAGE = 'IAM identity change audit is immutable';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_forbid_identity_change_mutation_v1() FROM PUBLIC;

CREATE TABLE iam_identity_change_audit (
    operation_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    resource_kind text COLLATE "C" NOT NULL,
    aggregate_id uuid NOT NULL,
    request_sha256 bytea NOT NULL,
    command_key text COLLATE "C" NOT NULL,
    result_code text COLLATE "C" NOT NULL,
    aggregate_revision bigint NOT NULL,
    authorization_clock bigint NOT NULL,
    event_sequence bigint NOT NULL,
    recorded_at timestamptz NOT NULL,
    CONSTRAINT iam_identity_change_audit_pkey PRIMARY KEY (operation_id),
    CONSTRAINT iam_identity_change_audit_receipt_fk FOREIGN KEY (operation_id)
        REFERENCES iam_authorization_operation_receipts (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_identity_change_audit_actor_fk FOREIGN KEY (actor_id)
        REFERENCES iam_users (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_identity_change_audit_stream_key UNIQUE (
        resource_kind, aggregate_id, event_sequence
    ),
    CONSTRAINT iam_identity_change_audit_resource_check CHECK (
        resource_kind IN ('user', 'role', 'organization')
    ),
    CONSTRAINT iam_identity_change_audit_request_check CHECK (
        pg_catalog.octet_length(request_sha256) = 32
    ),
    CONSTRAINT iam_identity_change_audit_command_check CHECK (
        command_key IN (
            'iam.identity.user.create', 'iam.identity.user.update',
            'iam.identity.user.disable', 'iam.identity.user.delete',
            'iam.identity.role.create', 'iam.identity.role.update',
            'iam.identity.role.retire', 'iam.identity.organization.create',
            'iam.identity.organization.update', 'iam.identity.organization.disable',
            'iam.identity.organization.delete'
        )
    ),
    CONSTRAINT iam_identity_change_audit_result_check CHECK (
        result_code = command_key || '.succeeded'
    ),
    CONSTRAINT iam_identity_change_audit_revision_check CHECK (
        aggregate_revision > 0 AND authorization_clock > 0 AND event_sequence > 0
    )
);

CREATE TRIGGER iam_identity_change_audit_mutation_guard
BEFORE UPDATE OR DELETE OR TRUNCATE ON iam_identity_change_audit
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_identity_change_mutation_v1();
ALTER TABLE iam_identity_change_audit
    ENABLE ALWAYS TRIGGER iam_identity_change_audit_mutation_guard;
