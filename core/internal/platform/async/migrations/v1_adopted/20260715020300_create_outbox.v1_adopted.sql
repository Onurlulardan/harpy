-- +goose Up
CREATE TABLE platform.outbox (
    event_id uuid PRIMARY KEY,
    tenant_id uuid,
    scope_kind text NOT NULL CHECK (scope_kind IN ('global', 'tenant')),
    scope_id uuid NOT NULL,
    authority_group_id text NOT NULL
        CHECK (authority_group_id ~ '^[a-z][a-z0-9_]{0,62}$'),
    authority_group_generation bigint NOT NULL CHECK (authority_group_generation > 0),
    authority_domain text NOT NULL
        CHECK (authority_domain ~ '^[a-z][a-z0-9_]{0,62}$'),
    cutover_generation bigint NOT NULL CHECK (cutover_generation > 0),
    writer_origin text NOT NULL CHECK (writer_origin IN ('v1', 'v2')),
    envelope_contract_version text NOT NULL
        CHECK (envelope_contract_version = 'integration-event.v1'),
    event_type text NOT NULL CHECK (event_type ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    event_version integer NOT NULL CHECK (event_version > 0),
    aggregate_type text NOT NULL
        CHECK (aggregate_type ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    aggregate_id uuid NOT NULL,
    aggregate_version bigint NOT NULL CHECK (aggregate_version > 0),
    event_sequence bigint NOT NULL CHECK (event_sequence > 0),
    occurred_at timestamptz NOT NULL,
    actor_id uuid NOT NULL,
    correlation_id text NOT NULL CHECK (
        correlation_id ~ '^[A-Za-z0-9][A-Za-z0-9_.:-]{0,63}$'
    ),
    causation_id uuid NOT NULL,
    idempotency_key text NOT NULL
        CHECK (pg_catalog.octet_length(idempotency_key) BETWEEN 1 AND 512),
    business_effect_idempotency_key text NOT NULL CHECK (
        pg_catalog.octet_length(business_effect_idempotency_key) BETWEEN 1 AND 512
    ),
    payload_canonical bytea,
    payload_sha256 text NOT NULL CHECK (payload_sha256 ~ '^[0-9a-f]{64}$'),
    payload_schema_sha256 text NOT NULL
        CHECK (payload_schema_sha256 ~ '^[0-9a-f]{64}$'),
    payload_redacted_at timestamptz,
    delivery_state text NOT NULL
        CHECK (delivery_state IN ('pending', 'publishing', 'delivered')),
    delivery_attempts integer NOT NULL CHECK (delivery_attempts >= 0),
    lease_owner_id uuid,
    lease_generation bigint NOT NULL CHECK (lease_generation >= 0),
    lease_heartbeat_at timestamptz,
    lease_until timestamptz,
    next_attempt_at timestamptz NOT NULL,
    confirmed_at timestamptz,
    delivered_at timestamptz,
    confirm_sha256 text CHECK (confirm_sha256 ~ '^[0-9a-f]{64}$'),
    last_error_code text CHECK (last_error_code ~ '^[A-Z][A-Z0-9_]{1,63}$'),
    last_error_sha256 text CHECK (last_error_sha256 ~ '^[0-9a-f]{64}$'),
    created_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    CONSTRAINT outbox_tenant_fk FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id) ON DELETE RESTRICT,
    CONSTRAINT outbox_authority_fk
        FOREIGN KEY (scope_kind, scope_id, authority_domain)
        REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON DELETE RESTRICT,
    CONSTRAINT outbox_authority_group_fk
        FOREIGN KEY (scope_kind, scope_id, authority_group_id)
        REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON DELETE RESTRICT,
    CONSTRAINT outbox_authority_member_fk
        FOREIGN KEY (scope_kind, scope_id, authority_group_id, authority_domain)
        REFERENCES platform.cutover_group_members (
            scope_kind, scope_id, group_id, domain
        ) ON DELETE RESTRICT,
    CONSTRAINT outbox_stream_sequence_unique UNIQUE (
        scope_kind, scope_id, aggregate_type, aggregate_id, event_sequence
    ),
    CONSTRAINT outbox_scope_check CHECK (
        (scope_kind = 'tenant' AND tenant_id IS NOT NULL AND tenant_id = scope_id) OR
        (scope_kind = 'global' AND tenant_id IS NULL)
    ),
    CONSTRAINT outbox_payload_check CHECK (
        (payload_canonical IS NOT NULL AND
            pg_catalog.octet_length(payload_canonical) BETWEEN 2 AND 65536 AND
            pg_catalog.jsonb_typeof(
                pg_catalog.convert_from(payload_canonical, 'UTF8')::jsonb
            ) = 'object' AND
            payload_sha256 = pg_catalog.encode(
                pg_catalog.sha256(payload_canonical), 'hex'
            ) AND
            payload_redacted_at IS NULL) OR
        (payload_canonical IS NULL AND payload_redacted_at IS NOT NULL AND
            delivery_state = 'delivered')
    ),
    CONSTRAINT outbox_attempt_check CHECK (
        lease_generation = delivery_attempts
    ),
    CONSTRAINT outbox_lease_check CHECK (
        (delivery_state = 'publishing' AND delivery_attempts > 0 AND
            lease_owner_id IS NOT NULL AND lease_heartbeat_at IS NOT NULL AND
            lease_until IS NOT NULL AND lease_heartbeat_at < lease_until) OR
        (delivery_state <> 'publishing' AND lease_owner_id IS NULL AND
            lease_heartbeat_at IS NULL AND lease_until IS NULL)
    ),
    CONSTRAINT outbox_delivery_check CHECK (
        (delivery_state IN ('pending', 'publishing') AND confirmed_at IS NULL AND
            delivered_at IS NULL AND confirm_sha256 IS NULL) OR
        (delivery_state = 'delivered' AND confirmed_at IS NOT NULL AND
            delivered_at IS NOT NULL AND confirm_sha256 IS NOT NULL)
    ),
    CONSTRAINT outbox_error_check CHECK (
        (last_error_code IS NULL AND last_error_sha256 IS NULL) OR
        (last_error_code IS NOT NULL AND last_error_sha256 IS NOT NULL)
    ),
    CONSTRAINT outbox_time_check CHECK (
        occurred_at <= created_at AND created_at <= updated_at AND
        (payload_redacted_at IS NULL OR
            (delivered_at IS NOT NULL AND delivered_at <= payload_redacted_at)) AND
        (confirmed_at IS NULL OR occurred_at <= confirmed_at) AND
        (delivered_at IS NULL OR confirmed_at <= delivered_at)
    )
);
