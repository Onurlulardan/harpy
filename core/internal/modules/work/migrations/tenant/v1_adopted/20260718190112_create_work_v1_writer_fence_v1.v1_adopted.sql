-- +goose Up
CREATE TABLE work_v1_writer_fence (
    singleton boolean PRIMARY KEY CHECK (singleton),
    state text NOT NULL CHECK (state IN ('V1_ACTIVE','V2_FENCED','V1_REOPENED','RETIRED')),
    run_id uuid,
    generation bigint CHECK (generation IS NULL OR generation > 0),
    reverse_reconciliation_id uuid,
    updated_by uuid,
    updated_at timestamptz NOT NULL,
    CHECK ((state = 'V1_ACTIVE' AND run_id IS NULL AND generation IS NULL AND
            reverse_reconciliation_id IS NULL AND updated_by IS NULL) OR
           (state = 'V2_FENCED' AND run_id IS NOT NULL AND generation IS NOT NULL AND
            reverse_reconciliation_id IS NULL AND updated_by IS NOT NULL) OR
           (state = 'V1_REOPENED' AND run_id IS NOT NULL AND generation IS NOT NULL AND
            reverse_reconciliation_id IS NOT NULL AND updated_by IS NOT NULL) OR
           (state = 'RETIRED' AND run_id IS NOT NULL AND generation IS NOT NULL AND
            reverse_reconciliation_id IS NULL AND updated_by IS NOT NULL))
);

INSERT INTO work_v1_writer_fence (
    singleton, state, run_id, generation, reverse_reconciliation_id, updated_by, updated_at
) VALUES (true, 'V1_ACTIVE', NULL, NULL, NULL, NULL, transaction_timestamp());
