-- +goose Up
CREATE TABLE iam_lifecycle_delivery_worker_claims (
    artifact_id uuid NOT NULL PRIMARY KEY,
    effect_id uuid NOT NULL,
    effect_idempotency_key text NOT NULL,
    state text NOT NULL,
    outcome_unknown_at timestamptz,
    artifact_ref text NOT NULL,
    message_id uuid NOT NULL,
    claimed_at timestamptz
);

-- +goose Down
DROP TABLE iam_lifecycle_delivery_worker_claims;
