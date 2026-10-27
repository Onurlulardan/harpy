-- +goose Up
CREATE TABLE platform.async_v1_adoption_runs (
    run_id uuid PRIMARY KEY,
    scope_kind text NOT NULL CHECK (scope_kind IN ('global', 'tenant')),
    scope_id uuid NOT NULL,
    domain text NOT NULL CHECK (domain ~ '^[a-z][a-z0-9_]{0,62}$'),
    group_id text NOT NULL CHECK (group_id ~ '^[a-z][a-z0-9_]{0,62}$'),
    source_generation bigint NOT NULL CHECK (source_generation > 0),
    target_generation bigint NOT NULL CHECK (target_generation = source_generation + 1),
    source_group_generation bigint NOT NULL CHECK (source_group_generation > 0),
    target_group_generation bigint NOT NULL
        CHECK (target_group_generation = source_group_generation + 1),
    inventory_contract_version text NOT NULL
        CHECK (inventory_contract_version = 'async-adoption-inventory.v1'),
    inventory_revision integer NOT NULL CHECK (inventory_revision = 1),
    inventory_sha256 text NOT NULL CHECK (inventory_sha256 ~ '^[0-9a-f]{64}$'),
    component_manifest_sha256 text NOT NULL
        CHECK (component_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    expected_enqueue_fences smallint NOT NULL
        CHECK (expected_enqueue_fences BETWEEN 1 AND 64),
    expected_scheduler_fences smallint NOT NULL
        CHECK (expected_scheduler_fences BETWEEN 1 AND 64),
    expected_consumer_fences smallint NOT NULL
        CHECK (expected_consumer_fences BETWEEN 1 AND 64),
    item_count bigint NOT NULL CHECK (item_count BETWEEN 0 AND 100000),
    state text NOT NULL CHECK (state IN ('inventory', 'disposing', 'accepted')),
    captured_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    accepted_at timestamptz,
    acceptance_sha256 text CHECK (acceptance_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT async_v1_runs_domain_fk
        FOREIGN KEY (scope_kind, scope_id, domain)
        REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON DELETE RESTRICT,
    CONSTRAINT async_v1_runs_group_fk
        FOREIGN KEY (scope_kind, scope_id, group_id)
        REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON DELETE RESTRICT,
    CONSTRAINT async_v1_runs_member_fk
        FOREIGN KEY (scope_kind, scope_id, group_id, domain)
        REFERENCES platform.cutover_group_members (
            scope_kind, scope_id, group_id, domain
        ) ON DELETE RESTRICT,
    CONSTRAINT async_v1_runs_authority_unique
        UNIQUE (scope_kind, scope_id, domain, source_generation, target_generation),
    CONSTRAINT async_v1_runs_terminal_check CHECK (
        (state IN ('inventory', 'disposing') AND accepted_at IS NULL AND
            acceptance_sha256 IS NULL) OR
        (state = 'accepted' AND accepted_at IS NOT NULL AND
            acceptance_sha256 IS NOT NULL)
    ),
    CONSTRAINT async_v1_runs_time_check CHECK (
        captured_at <= created_at AND (accepted_at IS NULL OR created_at <= accepted_at)
    )
);

CREATE TABLE platform.async_v1_adoption_items (
    run_id uuid NOT NULL REFERENCES platform.async_v1_adoption_runs (run_id)
        ON DELETE RESTRICT,
    inventory_class text NOT NULL CHECK (
        inventory_class IN ('queue', 'timer', 'future', 'running')
    ),
    source_surface text NOT NULL CHECK (source_surface IN (
        'queue', 'scheduler', 'delayed_plugin', 'ttl_queue',
        'future_state', 'running_state'
    )),
    source_name text NOT NULL CHECK (source_name ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    source_item_id text NOT NULL
        CHECK (pg_catalog.octet_length(source_item_id) BETWEEN 1 AND 512),
    source_contract_version text NOT NULL
        CHECK (source_contract_version ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    source_revision integer NOT NULL CHECK (source_revision > 0),
    source_payload_sha256 text NOT NULL
        CHECK (source_payload_sha256 ~ '^[0-9a-f]{64}$'),
    transform_version text NOT NULL
        CHECK (transform_version ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    idempotency_identity text NOT NULL
        CHECK (pg_catalog.octet_length(idempotency_identity) BETWEEN 1 AND 512),
    target_kind text NOT NULL CHECK (target_kind ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    target_payload_revision integer NOT NULL CHECK (target_payload_revision > 0),
    target_payload_ref text NOT NULL
        CHECK (pg_catalog.octet_length(target_payload_ref) BETWEEN 1 AND 2048),
    target_payload_sha256 text NOT NULL
        CHECK (target_payload_sha256 ~ '^[0-9a-f]{64}$'),
    target_run_at timestamptz NOT NULL,
    target_priority integer NOT NULL,
    target_max_attempts integer NOT NULL CHECK (target_max_attempts BETWEEN 1 AND 25),
    target_effect_class text NOT NULL
        CHECK (target_effect_class IN ('database_transactional', 'external')),
    target_effect_idempotency_key text
        CHECK (pg_catalog.octet_length(target_effect_idempotency_key) BETWEEN 1 AND 512),
    captured_at timestamptz NOT NULL,
    item_sha256 text NOT NULL CHECK (item_sha256 ~ '^[0-9a-f]{64}$'),
    PRIMARY KEY (run_id, inventory_class, source_name, source_item_id),
    CONSTRAINT async_v1_items_hash_unique UNIQUE (run_id, item_sha256),
    CONSTRAINT async_v1_items_class_surface_check CHECK (
        (inventory_class = 'queue' AND source_surface = 'queue') OR
        (inventory_class = 'timer' AND source_surface IN (
            'scheduler', 'delayed_plugin', 'ttl_queue'
        )) OR
        (inventory_class = 'future' AND source_surface = 'future_state') OR
        (inventory_class = 'running' AND source_surface = 'running_state')
    ),
    CONSTRAINT async_v1_items_effect_check CHECK (
        (target_effect_class = 'database_transactional' AND
            target_effect_idempotency_key IS NULL) OR
        (target_effect_class = 'external' AND
            target_effect_idempotency_key = idempotency_identity)
    )
);

CREATE INDEX async_v1_adoption_items_identity_idx
    ON platform.async_v1_adoption_items (run_id, idempotency_identity);
