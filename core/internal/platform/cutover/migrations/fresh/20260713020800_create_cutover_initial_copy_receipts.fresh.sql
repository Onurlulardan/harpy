-- +goose Up
CREATE TABLE platform.cutover_initial_copy_receipts (
    receipt_id uuid NOT NULL,
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    manifest_sha256 text NOT NULL,
    snapshot_name text NOT NULL,
    snapshot_lsn pg_lsn NOT NULL,
    relation_count integer NOT NULL,
    partition_count integer NOT NULL,
    row_count bigint NOT NULL,
    root_sha256 text NOT NULL,
    completed_at timestamptz NOT NULL,
    PRIMARY KEY (operation_id, run_id, manifest_sha256),
    CONSTRAINT cutover_initial_copy_receipts_id_unique UNIQUE (receipt_id),
    CONSTRAINT cutover_initial_copy_receipts_run_unique UNIQUE (operation_id, run_id),
    CONSTRAINT cutover_initial_copy_receipts_operation_fk
        FOREIGN KEY (operation_id, run_id)
        REFERENCES platform.cutover_operations (operation_id, run_id)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_initial_copy_receipts_manifest_check
        CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT cutover_initial_copy_receipts_snapshot_check
        CHECK (pg_catalog.octet_length(snapshot_name) BETWEEN 1 AND 256),
    CONSTRAINT cutover_initial_copy_receipts_count_check
        CHECK (relation_count > 0 AND partition_count > 0 AND row_count >= 0),
    CONSTRAINT cutover_initial_copy_receipts_root_check
        CHECK (root_sha256 ~ '^[0-9a-f]{64}$')
);

CREATE INDEX cutover_initial_copy_receipts_completed_idx
    ON platform.cutover_initial_copy_receipts (completed_at, operation_id, run_id);
