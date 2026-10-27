-- +goose Up
CREATE TABLE iam_platform_owner_transfer_audit (
    operation_id uuid PRIMARY KEY,
    predecessor_user_id uuid NOT NULL,
    successor_user_id uuid NOT NULL,
    request_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(request_sha256) = 32),
    owner_policy_revision bigint NOT NULL CHECK (owner_policy_revision > 1),
    platform_clock bigint NOT NULL CHECK (platform_clock > 1),
    head_generation bigint NOT NULL CHECK (head_generation > 1),
    recorded_at timestamptz NOT NULL,
    CONSTRAINT iam_owner_transfer_audit_predecessor_fk FOREIGN KEY (predecessor_user_id)
        REFERENCES iam_users (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_owner_transfer_audit_successor_fk FOREIGN KEY (successor_user_id)
        REFERENCES iam_users (id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT iam_owner_transfer_audit_distinct_ck CHECK (predecessor_user_id <> successor_user_id)
);

CREATE TABLE iam_platform_owner_transfer_outbox (
    event_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL UNIQUE,
    event_kind text COLLATE "C" NOT NULL CHECK (event_kind = 'iam.platform.owner_transferred.v1'),
    payload_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(payload_sha256) = 32),
    delivery_state text NOT NULL CHECK (delivery_state = 'pending'),
    occurred_at timestamptz NOT NULL,
    CONSTRAINT iam_owner_transfer_outbox_audit_fk FOREIGN KEY (operation_id)
        REFERENCES iam_platform_owner_transfer_audit (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);

-- +goose StatementBegin
CREATE FUNCTION iam_owner_transfer_observability_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'Owner transfer evidence is immutable';
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_owner_transfer_observability_guard_v1() FROM PUBLIC;

CREATE TRIGGER iam_owner_transfer_audit_guard
BEFORE UPDATE OR DELETE OR TRUNCATE ON iam_platform_owner_transfer_audit
FOR EACH STATEMENT EXECUTE FUNCTION iam_owner_transfer_observability_guard_v1();
ALTER TABLE iam_platform_owner_transfer_audit ENABLE ALWAYS TRIGGER iam_owner_transfer_audit_guard;

CREATE TRIGGER iam_owner_transfer_outbox_guard
BEFORE UPDATE OR DELETE OR TRUNCATE ON iam_platform_owner_transfer_outbox
FOR EACH STATEMENT EXECUTE FUNCTION iam_owner_transfer_observability_guard_v1();
ALTER TABLE iam_platform_owner_transfer_outbox ENABLE ALWAYS TRIGGER iam_owner_transfer_outbox_guard;
