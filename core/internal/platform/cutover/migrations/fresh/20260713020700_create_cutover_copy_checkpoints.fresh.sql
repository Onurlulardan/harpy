-- +goose Up
CREATE TABLE platform.cutover_copy_checkpoints (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    relation_ordinal smallint NOT NULL,
    checkpoint_key text NOT NULL,
    manifest_sha256 text NOT NULL,
    snapshot_name text NOT NULL,
    snapshot_lsn pg_lsn NOT NULL,
    last_key jsonb,
    received_lsn pg_lsn NOT NULL,
    applied_lsn pg_lsn NOT NULL,
    final_lsn pg_lsn,
    applied_rows bigint NOT NULL,
    applied_bytes bigint NOT NULL,
    attempt integer NOT NULL,
    state text NOT NULL,
    blocked_reason text,
    updated_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    PRIMARY KEY (operation_id, run_id, relation_ordinal, checkpoint_key),
    CONSTRAINT cutover_copy_checkpoints_relation_fk
        FOREIGN KEY (operation_id, run_id, relation_ordinal)
        REFERENCES platform.cutover_operation_relations (operation_id, run_id, ordinal)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_copy_checkpoints_key_fk
        FOREIGN KEY (operation_id, run_id, relation_ordinal, checkpoint_key)
        REFERENCES platform.cutover_operation_checkpoint_keys (
            operation_id, run_id, relation_ordinal, checkpoint_key
        ) ON DELETE RESTRICT,
    CONSTRAINT cutover_copy_checkpoints_key_check
        CHECK (pg_catalog.octet_length(checkpoint_key) BETWEEN 1 AND 512),
    CONSTRAINT cutover_copy_checkpoints_manifest_check
        CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT cutover_copy_checkpoints_snapshot_check
        CHECK (pg_catalog.octet_length(snapshot_name) BETWEEN 1 AND 256),
    CONSTRAINT cutover_copy_checkpoints_progress_check
        CHECK (
            applied_rows >= 0 AND applied_bytes >= 0 AND attempt > 0 AND
            applied_lsn <= received_lsn
        ),
    CONSTRAINT cutover_copy_checkpoints_state_check
        CHECK (state IN ('initial_copy', 'delta', 'blocked', 'complete')),
    CONSTRAINT cutover_copy_checkpoints_final_check
        CHECK (state <> 'complete' OR (final_lsn IS NOT NULL AND applied_lsn >= final_lsn)),
    CONSTRAINT cutover_copy_checkpoints_blocked_check
        CHECK (
            (state = 'blocked' AND pg_catalog.octet_length(blocked_reason) BETWEEN 1 AND 2048) OR
            (state <> 'blocked' AND blocked_reason IS NULL)
        )
);

CREATE INDEX cutover_copy_checkpoints_state_idx
    ON platform.cutover_copy_checkpoints (state, updated_at, operation_id, run_id);
