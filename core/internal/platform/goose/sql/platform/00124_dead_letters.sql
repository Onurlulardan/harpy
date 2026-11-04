-- +goose Up
CREATE TABLE platform.dead_letters (
    id uuid PRIMARY KEY,
    root_id uuid NOT NULL,
    entry_sequence integer NOT NULL,
    record_kind text NOT NULL,
    tenant_id uuid,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    authority_group_id text NOT NULL,
    authority_group_generation bigint NOT NULL,
    authority_domain text NOT NULL,
    cutover_generation bigint NOT NULL,
    writer_origin text NOT NULL,
    source_kind text NOT NULL,
    source_id uuid NOT NULL,
    consumer_name text,
    consumer_version integer,
    source_contract text NOT NULL,
    source_version integer NOT NULL,
    source_sha256 text NOT NULL,
    reason_class text NOT NULL,
    reason_code text NOT NULL,
    redacted_error_sha256 text NOT NULL,
    payload_canonical bytea,
    payload_sha256 text NOT NULL,
    payload_redacted_at timestamptz,
    redacted_envelope jsonb NOT NULL,
    full_payload_expires_at timestamptz NOT NULL,
    metadata_expires_at timestamptz NOT NULL,
    recorded_at timestamptz NOT NULL,
    action_reason text,
    action_ticket text,
    action_actor_id uuid,
    action_causation_id uuid
);

-- +goose Down
DROP TABLE platform.dead_letters;
