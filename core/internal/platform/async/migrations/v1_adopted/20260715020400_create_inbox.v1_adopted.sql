-- +goose Up
CREATE TABLE platform.inbox (
    id uuid PRIMARY KEY,
    record_kind text NOT NULL CHECK (record_kind IN ('event_receipt', 'watermark')),
    consumer_name text NOT NULL
        CHECK (consumer_name ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    consumer_version integer NOT NULL CHECK (consumer_version > 0),
    tenant_id uuid,
    scope_kind text NOT NULL CHECK (scope_kind IN ('global', 'tenant')),
    scope_id uuid NOT NULL,
    producer_group_id text NOT NULL
        CHECK (producer_group_id ~ '^[a-z][a-z0-9_]{0,62}$'),
    producer_group_generation bigint NOT NULL CHECK (producer_group_generation > 0),
    producer_domain text NOT NULL
        CHECK (producer_domain ~ '^[a-z][a-z0-9_]{0,62}$'),
    producer_generation bigint NOT NULL CHECK (producer_generation > 0),
    writer_origin text NOT NULL CHECK (writer_origin IN ('v1', 'v2')),
    consumer_group_id text NOT NULL
        CHECK (consumer_group_id ~ '^[a-z][a-z0-9_]{0,62}$'),
    consumer_group_generation bigint NOT NULL
        CHECK (consumer_group_generation > 0),
    consumer_domain text NOT NULL
        CHECK (consumer_domain ~ '^[a-z][a-z0-9_]{0,62}$'),
    consumer_generation bigint NOT NULL CHECK (consumer_generation > 0),
    consumer_writer_origin text NOT NULL CHECK (consumer_writer_origin = 'v2'),
    aggregate_type text NOT NULL
        CHECK (aggregate_type ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    aggregate_id uuid NOT NULL,
    event_id uuid,
    event_type text CHECK (event_type ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    event_version integer CHECK (event_version > 0),
    event_sequence bigint CHECK (event_sequence > 0),
    through_sequence bigint CHECK (through_sequence >= 0),
    watermark_kind text CHECK (watermark_kind IN ('event', 'bootstrap')),
    causation_id uuid,
    state_sha256 text CHECK (state_sha256 ~ '^[0-9a-f]{64}$'),
    receipt_sha256 text NOT NULL CHECK (receipt_sha256 ~ '^[0-9a-f]{64}$'),
    recorded_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    CONSTRAINT inbox_tenant_fk FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id) ON DELETE RESTRICT,
    CONSTRAINT inbox_authority_fk
        FOREIGN KEY (scope_kind, scope_id, producer_domain)
        REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON DELETE RESTRICT,
    CONSTRAINT inbox_authority_group_fk
        FOREIGN KEY (scope_kind, scope_id, producer_group_id)
        REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON DELETE RESTRICT,
    CONSTRAINT inbox_authority_member_fk
        FOREIGN KEY (scope_kind, scope_id, producer_group_id, producer_domain)
        REFERENCES platform.cutover_group_members (
            scope_kind, scope_id, group_id, domain
        ) ON DELETE RESTRICT,
    CONSTRAINT inbox_consumer_authority_fk
        FOREIGN KEY (scope_kind, scope_id, consumer_domain)
        REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON DELETE RESTRICT,
    CONSTRAINT inbox_consumer_authority_group_fk
        FOREIGN KEY (scope_kind, scope_id, consumer_group_id)
        REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON DELETE RESTRICT,
    CONSTRAINT inbox_consumer_authority_member_fk
        FOREIGN KEY (scope_kind, scope_id, consumer_group_id, consumer_domain)
        REFERENCES platform.cutover_group_members (
            scope_kind, scope_id, group_id, domain
        ) ON DELETE RESTRICT,
    CONSTRAINT inbox_event_unique UNIQUE (
        consumer_name, consumer_version, event_id
    ),
    CONSTRAINT inbox_event_sequence_unique UNIQUE (
        consumer_name, consumer_version, scope_kind, scope_id,
        aggregate_type, aggregate_id, event_sequence
    ),
    CONSTRAINT inbox_watermark_unique UNIQUE (
        consumer_name, consumer_version, scope_kind, scope_id,
        aggregate_type, aggregate_id, through_sequence
    ),
    CONSTRAINT inbox_scope_check CHECK (
        (scope_kind = 'tenant' AND tenant_id IS NOT NULL AND tenant_id = scope_id) OR
        (scope_kind = 'global' AND tenant_id IS NULL)
    ),
    CONSTRAINT inbox_record_shape_check CHECK (
        (record_kind = 'event_receipt' AND event_id IS NOT NULL AND
            event_type IS NOT NULL AND event_version IS NOT NULL AND
            event_sequence IS NOT NULL AND through_sequence IS NULL AND
            watermark_kind IS NULL AND causation_id IS NULL AND
            state_sha256 IS NULL) OR
        (record_kind = 'watermark' AND event_id IS NULL AND event_type IS NULL AND
            event_version IS NULL AND event_sequence IS NULL AND
            through_sequence IS NOT NULL AND watermark_kind IS NOT NULL)
    ),
    CONSTRAINT inbox_watermark_shape_check CHECK (
        record_kind <> 'watermark' OR
        (watermark_kind = 'event' AND causation_id IS NOT NULL AND
            state_sha256 IS NULL) OR
        (watermark_kind = 'bootstrap' AND causation_id IS NOT NULL AND
            state_sha256 IS NOT NULL)
    )
);
