-- +goose Up
CREATE TABLE platform.billing_subscription_audit (
    id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    source_operation_id uuid NOT NULL,
    source_kind text NOT NULL,
    kind text NOT NULL,
    actor_id uuid NOT NULL,
    before_revision bigint NOT NULL,
    after_revision bigint NOT NULL,
    occurred_at timestamptz,
    reason text NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_subscription_audit;
