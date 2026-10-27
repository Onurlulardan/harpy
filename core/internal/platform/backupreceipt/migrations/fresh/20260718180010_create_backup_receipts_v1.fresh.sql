-- +goose Up
CREATE TABLE platform.backup_receipts (
    receipt_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    package_id uuid NOT NULL,
    catalog_operation_id uuid NOT NULL,
    cutover_generations jsonb NOT NULL CHECK (jsonb_typeof(cutover_generations) = 'array'),
    owner_manifest_root text NOT NULL CHECK (owner_manifest_root ~ '^[0-9a-f]{64}$'),
    database_snapshot_root text NOT NULL CHECK (database_snapshot_root ~ '^[0-9a-f]{64}$'),
    object_inventory_root text NOT NULL CHECK (object_inventory_root ~ '^[0-9a-f]{64}$'),
    encrypted_storage_reference bytea NOT NULL CHECK (octet_length(encrypted_storage_reference) > 0),
    key_revision text NOT NULL CHECK (length(key_revision) BETWEEN 1 AND 128),
    completed_at timestamptz NOT NULL,
    restore_verified_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    verifier_version text NOT NULL CHECK (length(verifier_version) BETWEEN 1 AND 64),
    legal_hold boolean NOT NULL DEFAULT false,
    UNIQUE (tenant_id, package_id, catalog_operation_id),
    CHECK (date_trunc('microseconds', completed_at) = completed_at),
    CHECK (date_trunc('microseconds', restore_verified_at) = restore_verified_at),
    CHECK (date_trunc('microseconds', expires_at) = expires_at),
    CHECK (restore_verified_at >= completed_at),
    CHECK (expires_at > restore_verified_at)
);

-- +goose StatementBegin
CREATE FUNCTION platform.backup_reject_receipt_mutation_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY INVOKER
SET search_path = pg_catalog
AS $$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '55000',
        MESSAGE = 'backup receipt is immutable';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER backup_receipts_immutable
BEFORE UPDATE OR DELETE ON platform.backup_receipts
FOR EACH STATEMENT EXECUTE FUNCTION platform.backup_reject_receipt_mutation_v1();

GRANT SELECT ON TABLE platform.backup_receipts TO harpy_runtime;
