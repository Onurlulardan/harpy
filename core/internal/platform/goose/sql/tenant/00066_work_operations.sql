-- +goose Up
CREATE TABLE work_operations (
    state text NOT NULL PRIMARY KEY,
    aggregate_id uuid NOT NULL,
    aggregate_version bigint NOT NULL,
    response_status text NOT NULL,
    response_document jsonb NOT NULL,
    completed_at timestamptz,
    operation_id uuid NOT NULL,
    operation_kind text NOT NULL,
    actor_id uuid NOT NULL,
    idempotency_key text NOT NULL,
    request_sha256 text NOT NULL,
    aggregate_kind text NOT NULL,
    occurred_at timestamptz
);

-- +goose Down
DROP TABLE work_operations;
