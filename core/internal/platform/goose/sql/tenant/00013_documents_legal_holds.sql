-- +goose Up
CREATE TABLE documents_legal_holds (
    id uuid NOT NULL PRIMARY KEY,
    lifecycle_state text NOT NULL,
    source_owner text NOT NULL,
    source_case_id uuid NOT NULL,
    item_id uuid NOT NULL,
    evidence_sha256 text NOT NULL,
    placed_at timestamptz
);

-- +goose Down
DROP TABLE documents_legal_holds;
