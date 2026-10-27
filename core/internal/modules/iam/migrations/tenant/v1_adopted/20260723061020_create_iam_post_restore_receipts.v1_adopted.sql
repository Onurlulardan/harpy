-- +goose Up
CREATE TABLE iam_restored_identity_receipts (
    restore_operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    recovery_set_id uuid NOT NULL,
    owner_user_id uuid NOT NULL,
    anchor_id uuid NOT NULL,
    security_epoch bytea NOT NULL CHECK (pg_catalog.octet_length(security_epoch)=32),
    restore_receipt_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(restore_receipt_sha256)=32),
    protected_head_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(protected_head_sha256)=32),
    custody_generation bigint NOT NULL CHECK (custody_generation>0),
    custody_result_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(custody_result_sha256)=32),
    invalidation_counts jsonb NOT NULL CHECK (pg_catalog.jsonb_typeof(invalidation_counts)='object'),
    invalidation_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(invalidation_sha256)=32),
    created_at timestamptz NOT NULL
);

CREATE TABLE iam_post_restore_owner_receipts (
    restore_operation_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL UNIQUE,
    owner_user_id uuid NOT NULL,
    anchor_id uuid NOT NULL,
    assertion_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(assertion_sha256)=32),
    password_credential_sha256 bytea NOT NULL CHECK (
        pg_catalog.octet_length(password_credential_sha256)=32
    ),
    receipt_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(receipt_sha256)=32),
    created_at timestamptz NOT NULL,
    CONSTRAINT iam_post_restore_owner_user_fk FOREIGN KEY(owner_user_id)
        REFERENCES iam_users(id) ON UPDATE NO ACTION ON DELETE RESTRICT
);

-- +goose StatementBegin
CREATE FUNCTION iam_forbid_post_restore_receipt_mutation_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $function$
BEGIN
    RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='post-restore receipt is immutable';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_forbid_post_restore_receipt_mutation_v1() FROM PUBLIC;
CREATE TRIGGER iam_restored_identity_receipt_guard
BEFORE UPDATE OR DELETE OR TRUNCATE ON iam_restored_identity_receipts
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_post_restore_receipt_mutation_v1();
CREATE TRIGGER iam_post_restore_owner_receipt_guard
BEFORE UPDATE OR DELETE OR TRUNCATE ON iam_post_restore_owner_receipts
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_post_restore_receipt_mutation_v1();
ALTER TABLE iam_restored_identity_receipts
    ENABLE ALWAYS TRIGGER iam_restored_identity_receipt_guard;
ALTER TABLE iam_post_restore_owner_receipts
    ENABLE ALWAYS TRIGGER iam_post_restore_owner_receipt_guard;
