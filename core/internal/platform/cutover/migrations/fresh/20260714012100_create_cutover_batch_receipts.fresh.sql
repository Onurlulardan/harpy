-- +goose Up
CREATE TABLE platform.cutover_batch_receipts (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    relation_ordinal smallint NOT NULL,
    checkpoint_key text NOT NULL,
    batch_ordinal integer NOT NULL,
    source_xid xid8 NOT NULL,
    source_commit_lsn pg_lsn NOT NULL,
    batch_sha256 text NOT NULL,
    applied_count bigint NOT NULL,
    key_state_sha256 text NOT NULL,
    state_sha256 text NOT NULL,
    applied_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    PRIMARY KEY (operation_id, run_id, relation_ordinal, checkpoint_key, batch_ordinal),
    CONSTRAINT cutover_batch_receipts_relation_fk
        FOREIGN KEY (operation_id, run_id, relation_ordinal)
        REFERENCES platform.cutover_operation_relations (operation_id, run_id, ordinal)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_batch_receipts_checkpoint_fk
        FOREIGN KEY (operation_id, run_id, relation_ordinal, checkpoint_key)
        REFERENCES platform.cutover_operation_checkpoint_keys (
            operation_id, run_id, relation_ordinal, checkpoint_key
        ) ON DELETE RESTRICT,
    CONSTRAINT cutover_batch_receipts_checkpoint_check
        CHECK (pg_catalog.octet_length(checkpoint_key) BETWEEN 1 AND 512),
    CONSTRAINT cutover_batch_receipts_progress_check
        CHECK (batch_ordinal BETWEEN 1 AND 2147483 AND applied_count >= 0),
    CONSTRAINT cutover_batch_receipts_digest_check
        CHECK (
            batch_sha256 ~ '^[0-9a-f]{64}$' AND
            key_state_sha256 ~ '^[0-9a-f]{64}$' AND
            state_sha256 ~ '^[0-9a-f]{64}$'
        )
);

CREATE UNIQUE INDEX cutover_batch_receipts_source_idx
    ON platform.cutover_batch_receipts (
        operation_id, run_id, relation_ordinal, source_xid, batch_ordinal
    );
