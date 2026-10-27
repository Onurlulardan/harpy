-- +goose Up
CREATE TABLE platform.cutover_contract_probe_rows (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    relation_ordinal smallint NOT NULL,
    checkpoint_key text NOT NULL,
    batch_ordinal integer NOT NULL,
    batch_sha256 text NOT NULL,
    canonical_payload jsonb NOT NULL,
    applied_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    PRIMARY KEY (operation_id, run_id, relation_ordinal, checkpoint_key, batch_ordinal),
    CONSTRAINT cutover_contract_probe_rows_digest_check
        CHECK (batch_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT cutover_contract_probe_rows_payload_check
        CHECK (pg_catalog.octet_length(canonical_payload::text) BETWEEN 2 AND 8388608)
);
