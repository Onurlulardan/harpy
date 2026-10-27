-- +goose Up
CREATE TABLE platform.dead_letters (
    id uuid PRIMARY KEY,
    root_id uuid NOT NULL,
    entry_sequence integer NOT NULL CHECK (entry_sequence >= 0),
    record_kind text NOT NULL
        CHECK (record_kind IN ('root', 'retry_action', 'repair_action')),
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
    source_kind text NOT NULL CHECK (source_kind IN ('job', 'event')),
    source_id uuid NOT NULL,
    consumer_name text
        CHECK (consumer_name ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    consumer_version integer CHECK (consumer_version > 0),
    source_contract text NOT NULL
        CHECK (source_contract ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    source_version integer NOT NULL CHECK (source_version > 0),
    source_sha256 text NOT NULL CHECK (source_sha256 ~ '^[0-9a-f]{64}$'),
    reason_class text NOT NULL CHECK (reason_class IN (
        'permanent', 'unknown_version', 'delivery_limit',
        'outcome_unknown', 'stale_generation'
    )),
    reason_code text NOT NULL CHECK (reason_code ~ '^[A-Z][A-Z0-9_]{1,63}$'),
    redacted_error_sha256 text NOT NULL
        CHECK (redacted_error_sha256 ~ '^[0-9a-f]{64}$'),
    payload_canonical bytea CHECK (
        pg_catalog.octet_length(payload_canonical) BETWEEN 2 AND 65536
    ),
    payload_sha256 text NOT NULL CHECK (payload_sha256 ~ '^[0-9a-f]{64}$'),
    payload_redacted_at timestamptz,
    redacted_envelope jsonb NOT NULL,
    action_reason text,
    action_ticket text,
    action_actor_id uuid,
    action_causation_id uuid,
    full_payload_expires_at timestamptz NOT NULL,
    metadata_expires_at timestamptz NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    CONSTRAINT dead_letters_root_fk FOREIGN KEY (root_id)
        REFERENCES platform.dead_letters (id) ON DELETE RESTRICT,
    CONSTRAINT dead_letters_tenant_fk FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id) ON DELETE RESTRICT,
    CONSTRAINT dead_letters_authority_fk
        FOREIGN KEY (scope_kind, scope_id, authority_domain)
        REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON DELETE RESTRICT,
    CONSTRAINT dead_letters_authority_group_fk
        FOREIGN KEY (scope_kind, scope_id, authority_group_id)
        REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON DELETE RESTRICT,
    CONSTRAINT dead_letters_authority_member_fk
        FOREIGN KEY (scope_kind, scope_id, authority_group_id, authority_domain)
        REFERENCES platform.cutover_group_members (
            scope_kind, scope_id, group_id, domain
        ) ON DELETE RESTRICT,
    CONSTRAINT dead_letters_chain_unique UNIQUE (root_id, entry_sequence),
    CONSTRAINT dead_letters_scope_check CHECK (
        (scope_kind = 'tenant' AND tenant_id IS NOT NULL AND tenant_id = scope_id) OR
        (scope_kind = 'global' AND tenant_id IS NULL)
    ),
    CONSTRAINT dead_letters_source_shape_check CHECK (
        (source_kind = 'job' AND consumer_name IS NULL AND
            consumer_version IS NULL) OR
        (source_kind = 'event' AND consumer_name IS NOT NULL AND
            consumer_version IS NOT NULL)
    ),
    CONSTRAINT dead_letters_envelope_check CHECK (
        pg_catalog.jsonb_typeof(redacted_envelope) = 'object' AND
        pg_catalog.octet_length(redacted_envelope::text) <= 16384
    ),
    CONSTRAINT dead_letters_payload_hash_check CHECK (
        payload_canonical IS NULL OR (
            pg_catalog.jsonb_typeof(
                pg_catalog.convert_from(payload_canonical, 'UTF8')::jsonb
            ) = 'object' AND
            payload_sha256 = pg_catalog.encode(
                pg_catalog.sha256(payload_canonical), 'hex'
            )
        )
    ),
    CONSTRAINT dead_letters_root_shape_check CHECK (
        (record_kind = 'root' AND id = root_id AND entry_sequence = 0 AND
            action_reason IS NULL AND action_ticket IS NULL AND
            action_actor_id IS NULL AND action_causation_id IS NULL AND
            ((payload_canonical IS NOT NULL AND payload_redacted_at IS NULL) OR
             (payload_canonical IS NULL AND payload_redacted_at IS NOT NULL))) OR
        (record_kind IN ('retry_action', 'repair_action') AND id <> root_id AND
            entry_sequence > 0 AND payload_canonical IS NULL AND
            payload_redacted_at IS NULL AND
            action_reason IS NOT NULL AND action_ticket IS NOT NULL AND
            action_actor_id IS NOT NULL AND action_causation_id IS NOT NULL)
    ),
    CONSTRAINT dead_letters_action_text_check CHECK (
        (action_reason IS NULL AND action_ticket IS NULL) OR
        (action_reason IS NOT NULL AND action_ticket IS NOT NULL AND
            pg_catalog.octet_length(action_reason) BETWEEN 1 AND 2048 AND
            pg_catalog.octet_length(action_ticket) BETWEEN 1 AND 256)
    ),
    CONSTRAINT dead_letters_time_check CHECK (
        full_payload_expires_at <= metadata_expires_at AND
        recorded_at < metadata_expires_at AND
        (record_kind <> 'root' OR recorded_at < full_payload_expires_at) AND
        (payload_redacted_at IS NULL OR
            full_payload_expires_at <= payload_redacted_at)
    )
);
