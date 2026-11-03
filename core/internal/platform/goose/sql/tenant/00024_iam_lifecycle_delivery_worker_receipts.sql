-- +goose Up
CREATE TABLE iam_lifecycle_delivery_worker_receipts (
    receipt_id uuid NOT NULL PRIMARY KEY,
    artifact_id uuid NOT NULL,
    purpose text NOT NULL,
    subject_user_id uuid NOT NULL,
    resource_id uuid NOT NULL,
    token_id uuid NOT NULL,
    artifact_ref text NOT NULL,
    effect_id uuid NOT NULL,
    effect_idempotency_key text NOT NULL,
    message_id uuid NOT NULL,
    outcome text NOT NULL,
    code text NOT NULL,
    provider_reference text NOT NULL,
    result_sha256 text NOT NULL,
    completed_at timestamptz
);

-- +goose Down
DROP TABLE iam_lifecycle_delivery_worker_receipts;
