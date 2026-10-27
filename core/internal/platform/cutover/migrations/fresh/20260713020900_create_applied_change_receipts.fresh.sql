-- +goose Up
CREATE TABLE platform.applied_change_receipts (
    receipt_id uuid NOT NULL,
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    manifest_sha256 text NOT NULL,
    source_xid xid8 NOT NULL,
    source_commit_lsn pg_lsn NOT NULL,
    relation_ordinal smallint NOT NULL,
    change_ordinal integer NOT NULL,
    checkpoint_key text NOT NULL,
    mutation_kind text NOT NULL,
    mutation_sha256 text NOT NULL,
    applied_at timestamptz NOT NULL,
    PRIMARY KEY (
        operation_id,
        run_id,
        manifest_sha256,
        source_xid,
        relation_ordinal,
        change_ordinal
    ),
    CONSTRAINT applied_change_receipts_id_unique UNIQUE (receipt_id),
    CONSTRAINT applied_change_receipts_relation_fk
        FOREIGN KEY (operation_id, run_id, relation_ordinal)
        REFERENCES platform.cutover_operation_relations (operation_id, run_id, ordinal)
        ON DELETE RESTRICT,
    CONSTRAINT applied_change_receipts_checkpoint_fk
        FOREIGN KEY (operation_id, run_id, relation_ordinal, checkpoint_key)
        REFERENCES platform.cutover_operation_checkpoint_keys (
            operation_id, run_id, relation_ordinal, checkpoint_key
        ) ON DELETE RESTRICT,
    CONSTRAINT applied_change_receipts_manifest_check
        CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT applied_change_receipts_ordinal_check CHECK (change_ordinal > 0),
    CONSTRAINT applied_change_receipts_checkpoint_check
        CHECK (pg_catalog.octet_length(checkpoint_key) BETWEEN 1 AND 512),
    CONSTRAINT applied_change_receipts_kind_check
        CHECK (mutation_kind IN ('insert', 'update', 'delete')),
    CONSTRAINT applied_change_receipts_mutation_check
        CHECK (mutation_sha256 ~ '^[0-9a-f]{64}$')
);

CREATE INDEX applied_change_receipts_lsn_idx
    ON platform.applied_change_receipts (operation_id, run_id, source_commit_lsn);
