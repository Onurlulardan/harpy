-- +goose Up
CREATE TABLE platform.billing_provider_events (
    id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    webhook_receipt_id uuid NOT NULL,
    provider_object_mapping_id uuid NOT NULL,
    normalized_schema text NOT NULL,
    normalized_version bigint NOT NULL,
    ordering_mode text NOT NULL,
    ordering_sequence text NOT NULL,
    ordering_previous text NOT NULL,
    canonical_fact text NOT NULL,
    occurred_at timestamptz,
    normalized_at timestamptz,
    normalized_sha256 text NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_provider_events;
