-- +goose Up
CREATE TABLE iam_system_import_row_receipts (
    import_id uuid NOT NULL,
    job_id uuid NOT NULL,
    row_index bigint NOT NULL,
    outcome text NOT NULL,
    owner_kind text,
    owner_id uuid,
    error_code text,
    source_sha256 bytea NOT NULL,
    result_sha256 bytea NOT NULL,
    owner_revision bigint,
    committed_at timestamptz NOT NULL,
    PRIMARY KEY (import_id, row_index),
    CONSTRAINT iam_system_import_row_receipts_import_fk FOREIGN KEY (import_id)
        REFERENCES iam_system_record_imports (import_id) ON DELETE RESTRICT,
    CONSTRAINT iam_system_import_row_receipts_shape_check CHECK (
        pg_catalog.uuid_extract_version(import_id) = 7 AND
        pg_catalog.uuid_extract_version(job_id) = 7 AND
        row_index BETWEEN 1 AND 1000000 AND
        pg_catalog.octet_length(source_sha256) = 32 AND
        pg_catalog.octet_length(result_sha256) = 32 AND
        ((outcome IN ('created','updated') AND
          owner_kind IN ('record','invitation') AND owner_id IS NOT NULL AND
          error_code IS NULL AND owner_revision IS NOT NULL) OR
         (outcome = 'skipped' AND owner_kind IS NULL AND owner_id IS NULL AND
          error_code IS NULL AND owner_revision IS NULL) OR
         (outcome = 'rejected' AND owner_kind IS NULL AND owner_id IS NULL AND
          error_code ~ '^[A-Z][A-Z0-9_]{1,63}$' AND owner_revision IS NULL)) AND
        (owner_revision IS NULL OR owner_revision BETWEEN 1 AND 9007199254740991) AND
        pg_catalog.isfinite(committed_at)
    )
);

CREATE TABLE iam_system_transfer_admissions (
    transfer_kind text NOT NULL,
    transfer_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    state text NOT NULL,
    tenant_run_slot smallint,
    principal_wait_slot smallint,
    admitted_at timestamptz NOT NULL,
    released_at timestamptz,
    PRIMARY KEY (transfer_kind, transfer_id),
    CONSTRAINT iam_system_transfer_admissions_shape_check CHECK (
        transfer_kind IN ('export','import') AND
        pg_catalog.uuid_extract_version(transfer_id) = 7 AND
        tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid AND
        actor_id <> '00000000-0000-0000-0000-000000000000'::uuid AND
        ((state = 'waiting' AND tenant_run_slot IS NULL AND
          principal_wait_slot BETWEEN 1 AND 5 AND released_at IS NULL) OR
         (state = 'running' AND tenant_run_slot BETWEEN 1 AND 2 AND
          principal_wait_slot IS NULL AND released_at IS NULL) OR
         (state = 'released' AND tenant_run_slot IS NULL AND
          principal_wait_slot IS NULL AND released_at >= admitted_at)) AND
        pg_catalog.isfinite(admitted_at)
    )
);

CREATE UNIQUE INDEX iam_system_transfer_tenant_run_slot_uidx
ON iam_system_transfer_admissions (tenant_id, tenant_run_slot)
WHERE state = 'running';

CREATE UNIQUE INDEX iam_system_transfer_principal_wait_slot_uidx
ON iam_system_transfer_admissions (actor_id, principal_wait_slot)
WHERE state = 'waiting';

CREATE INDEX iam_system_import_receipts_report_idx
ON iam_system_import_row_receipts (import_id, row_index)
WHERE outcome = 'rejected';
