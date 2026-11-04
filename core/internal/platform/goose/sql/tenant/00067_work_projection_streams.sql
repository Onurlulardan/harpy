-- +goose Up
CREATE TABLE work_projection_streams (
    bootstrap_sequence text NOT NULL PRIMARY KEY,
    aggregate_id uuid NOT NULL,
    state text NOT NULL,
    through_sequence text NOT NULL,
    source_owner text NOT NULL,
    aggregate_kind text NOT NULL,
    state_sha256 text NOT NULL,
    updated_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE work_projection_streams;
