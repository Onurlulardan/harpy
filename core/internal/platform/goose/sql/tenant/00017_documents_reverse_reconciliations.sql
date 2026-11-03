-- +goose Up
CREATE TABLE documents_reverse_reconciliations (
    lifecycle_state text NOT NULL PRIMARY KEY,
    approved_by uuid NOT NULL,
    decided_at timestamptz,
    reconciliation_id uuid NOT NULL,
    source_generation bigint NOT NULL,
    target_generation bigint NOT NULL,
    source_count bigint NOT NULL,
    target_count bigint NOT NULL,
    source_sha256 text NOT NULL,
    target_sha256 text NOT NULL,
    requested_by uuid NOT NULL,
    created_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE documents_reverse_reconciliations;
