-- +goose Up
CREATE TABLE platform.inbox (
    id uuid NOT NULL,
    record_kind text NOT NULL,
    consumer_name text NOT NULL,
    consumer_version integer NOT NULL,
    tenant_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    producer_group_id text NOT NULL,
    producer_group_generation bigint NOT NULL,
    producer_domain text NOT NULL,
    producer_generation bigint NOT NULL,
    writer_origin text NOT NULL,
    consumer_group_id text NOT NULL,
    consumer_group_generation bigint NOT NULL,
    consumer_domain text NOT NULL,
    consumer_generation bigint NOT NULL,
    consumer_writer_origin text NOT NULL,
    aggregate_type text NOT NULL,
    aggregate_id uuid NOT NULL,
    through_sequence bigint,
    watermark_kind text,
    causation_id uuid,
    state_sha256 text,
    receipt_sha256 text NOT NULL,
    event_id uuid,
    event_type text,
    event_version integer,
    event_sequence bigint
);

-- +goose Down
DROP TABLE platform.inbox;
