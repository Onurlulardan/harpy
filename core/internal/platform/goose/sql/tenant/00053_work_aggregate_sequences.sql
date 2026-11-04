-- +goose Up
CREATE TABLE work_aggregate_sequences (
    aggregate_kind text NOT NULL PRIMARY KEY,
    aggregate_id uuid NOT NULL,
    event_sequence text NOT NULL,
    aggregate_version bigint NOT NULL,
    updated_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE work_aggregate_sequences;
