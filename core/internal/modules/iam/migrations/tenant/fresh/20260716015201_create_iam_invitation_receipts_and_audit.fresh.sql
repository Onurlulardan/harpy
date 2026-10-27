-- +goose Up
CREATE TABLE iam_invitation_operation_receipts (
    operation_id uuid NOT NULL,
    command_key text COLLATE "C" NOT NULL,
    actor_id uuid NOT NULL,
    request_sha256 bytea NOT NULL,
    result_id uuid NOT NULL,
    result_state text NOT NULL,
    result_revision bigint NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_invitation_operation_receipts_pkey PRIMARY KEY (operation_id),
    CONSTRAINT iam_invitation_operation_receipts_actor_fk
        FOREIGN KEY (actor_id) REFERENCES iam_users (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_invitation_operation_receipts_result_fk
        FOREIGN KEY (result_id) REFERENCES iam_invitations (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_invitation_operation_receipts_command_check CHECK (
        command_key IN ('iam.invitation.create', 'iam.invitation.revoke',
                        'iam.invitation.expire')
    ),
    CONSTRAINT iam_invitation_operation_receipts_result_check CHECK (
        pg_catalog.octet_length(request_sha256) = 32 AND
        result_state IN ('pending', 'revoked', 'expired') AND result_revision > 0
    )
);

CREATE TABLE iam_invitation_change_audit (
    operation_id uuid NOT NULL,
    command_key text COLLATE "C" NOT NULL,
    actor_id uuid NOT NULL,
    invitation_id uuid NOT NULL,
    request_sha256 bytea NOT NULL,
    result_state text NOT NULL,
    result_revision bigint NOT NULL,
    delivery_artifact_ref text COLLATE "C",
    authorization_clock bigint,
    recorded_at timestamptz NOT NULL,
    CONSTRAINT iam_invitation_change_audit_pkey PRIMARY KEY (operation_id),
    CONSTRAINT iam_invitation_change_audit_actor_fk FOREIGN KEY (actor_id)
        REFERENCES iam_users (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_invitation_change_audit_invitation_fk FOREIGN KEY (invitation_id)
        REFERENCES iam_invitations (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_invitation_change_audit_command_check CHECK (
        command_key IN ('iam.invitation.create', 'iam.invitation.revoke',
                        'iam.invitation.expire', 'iam.invitation.accept')
    ),
    CONSTRAINT iam_invitation_change_audit_result_check CHECK (
        pg_catalog.octet_length(request_sha256) = 32 AND result_revision > 0 AND
        ((command_key = 'iam.invitation.create' AND result_state = 'pending' AND
            delivery_artifact_ref IS NOT NULL AND authorization_clock IS NULL) OR
         (command_key IN ('iam.invitation.revoke', 'iam.invitation.expire') AND
            delivery_artifact_ref IS NULL AND authorization_clock IS NULL) OR
         (command_key = 'iam.invitation.accept' AND result_state = 'accepted' AND
            delivery_artifact_ref IS NULL AND authorization_clock > 0))
    )
);

CREATE TRIGGER iam_invitation_operation_receipts_mutation_guard
BEFORE UPDATE OR DELETE OR TRUNCATE ON iam_invitation_operation_receipts
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_invitation_operation_receipts ENABLE ALWAYS TRIGGER
    iam_invitation_operation_receipts_mutation_guard;
CREATE TRIGGER iam_invitation_change_audit_mutation_guard
BEFORE UPDATE OR DELETE OR TRUNCATE ON iam_invitation_change_audit
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_invitation_change_audit ENABLE ALWAYS TRIGGER
    iam_invitation_change_audit_mutation_guard;
