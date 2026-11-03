-- +goose Up
CREATE TABLE documents_operations (
    operation_id uuid NOT NULL PRIMARY KEY,
    operation_kind text NOT NULL,
    actor_id uuid NOT NULL,
    idempotency_key text NOT NULL,
    request_sha256 text NOT NULL,
    aggregate_kind text NOT NULL,
    aggregate_id uuid NOT NULL,
    aggregate_version bigint NOT NULL,
    response_status text NOT NULL,
    response_document jsonb NOT NULL,
    occurred_at timestamptz
);

-- +goose Down
DROP TABLE documents_operations;
