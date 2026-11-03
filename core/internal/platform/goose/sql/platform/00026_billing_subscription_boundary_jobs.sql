-- +goose Up
CREATE TABLE platform.billing_subscription_boundary_jobs (
    id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    boundary_kind text NOT NULL,
    run_at timestamptz,
    expected_revision bigint NOT NULL,
    effect_id uuid NOT NULL,
    payload_revision bigint NOT NULL,
    payload_reference text NOT NULL,
    payload_sha256 text NOT NULL,
    created_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_subscription_boundary_jobs;
