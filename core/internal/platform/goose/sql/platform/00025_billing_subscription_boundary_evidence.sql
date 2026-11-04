-- +goose Up
CREATE TABLE platform.billing_subscription_boundary_evidence (
    id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    boundary_job_id uuid NOT NULL,
    provider_event_id uuid NOT NULL,
    boundary_kind text NOT NULL,
    due_at timestamptz,
    expected_revision bigint NOT NULL,
    outcome text NOT NULL,
    next_period_start text NOT NULL,
    next_period_end text NOT NULL,
    grace_ends_at timestamptz,
    recorded_at timestamptz
);

-- +goose Down
DROP TABLE platform.billing_subscription_boundary_evidence;
