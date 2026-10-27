-- +goose Up
CREATE TABLE platform.async_v1_adoption_dispositions (
    run_id uuid NOT NULL,
    inventory_class text NOT NULL,
    source_name text NOT NULL,
    source_item_id text NOT NULL,
    item_sha256 text NOT NULL CHECK (item_sha256 ~ '^[0-9a-f]{64}$'),
    idempotency_identity text NOT NULL
        CHECK (pg_catalog.octet_length(idempotency_identity) BETWEEN 1 AND 512),
    disposition text NOT NULL CHECK (disposition IN ('adopted', 'cancelled')),
    target_job_id uuid REFERENCES platform.jobs (id) ON DELETE RESTRICT,
    reason_code text CHECK (reason_code ~ '^[A-Z][A-Z0-9_]{1,63}$'),
    reason text CHECK (pg_catalog.octet_length(reason) BETWEEN 1 AND 2048),
    ticket text CHECK (pg_catalog.octet_length(ticket) BETWEEN 1 AND 256),
    disposed_at timestamptz NOT NULL,
    receipt_sha256 text NOT NULL CHECK (receipt_sha256 ~ '^[0-9a-f]{64}$'),
    PRIMARY KEY (run_id, inventory_class, source_name, source_item_id),
    CONSTRAINT async_v1_dispositions_item_fk FOREIGN KEY (
        run_id, inventory_class, source_name, source_item_id
    ) REFERENCES platform.async_v1_adoption_items (
        run_id, inventory_class, source_name, source_item_id
    ) ON DELETE RESTRICT,
    CONSTRAINT async_v1_dispositions_hash_unique UNIQUE (run_id, receipt_sha256),
    CONSTRAINT async_v1_dispositions_target_unique UNIQUE (run_id, target_job_id),
    CONSTRAINT async_v1_dispositions_outcome_check CHECK (
        (disposition = 'adopted' AND target_job_id IS NOT NULL AND
            reason_code IS NULL AND reason IS NULL AND ticket IS NULL) OR
        (disposition = 'cancelled' AND target_job_id IS NULL AND
            reason_code IS NOT NULL AND reason IS NOT NULL AND ticket IS NOT NULL)
    )
);

CREATE UNIQUE INDEX async_v1_dispositions_adopted_identity_unique
    ON platform.async_v1_adoption_dispositions (run_id, idempotency_identity)
    WHERE disposition = 'adopted';

CREATE TABLE platform.async_v1_adoption_job_bindings (
    run_id uuid NOT NULL,
    inventory_class text NOT NULL,
    source_name text NOT NULL,
    source_item_id text NOT NULL,
    item_sha256 text NOT NULL CHECK (item_sha256 ~ '^[0-9a-f]{64}$'),
    job_id uuid NOT NULL UNIQUE,
    bound_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    PRIMARY KEY (run_id, inventory_class, source_name, source_item_id),
    CONSTRAINT async_v1_job_bindings_item_fk FOREIGN KEY (
        run_id, inventory_class, source_name, source_item_id
    ) REFERENCES platform.async_v1_adoption_items (
        run_id, inventory_class, source_name, source_item_id
    ) ON DELETE RESTRICT
);

CREATE TABLE platform.async_v1_restart_fence_receipts (
    run_id uuid NOT NULL REFERENCES platform.async_v1_adoption_runs (run_id)
        ON DELETE RESTRICT,
    fence_kind text NOT NULL CHECK (fence_kind IN ('enqueue', 'scheduler', 'consumer')),
    component_name text NOT NULL
        CHECK (component_name ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    component_contract_version text NOT NULL
        CHECK (component_contract_version ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    binary_sha256 text NOT NULL CHECK (binary_sha256 ~ '^[0-9a-f]{64}$'),
    config_sha256 text NOT NULL CHECK (config_sha256 ~ '^[0-9a-f]{64}$'),
    restart_probe_sha256 text NOT NULL
        CHECK (restart_probe_sha256 ~ '^[0-9a-f]{64}$'),
    fenced_at timestamptz NOT NULL,
    restart_verified_at timestamptz NOT NULL,
    receipt_sha256 text NOT NULL CHECK (receipt_sha256 ~ '^[0-9a-f]{64}$'),
    recorded_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    PRIMARY KEY (run_id, fence_kind, component_name),
    CONSTRAINT async_v1_restart_fence_hash_unique UNIQUE (run_id, receipt_sha256),
    CONSTRAINT async_v1_restart_fence_time_check CHECK (
        fenced_at < restart_verified_at AND restart_verified_at <= recorded_at
    )
);

CREATE TABLE platform.async_v1_zero_inventory_receipts (
    run_id uuid PRIMARY KEY REFERENCES platform.async_v1_adoption_runs (run_id)
        ON DELETE RESTRICT,
    inventory_contract_version text NOT NULL
        CHECK (inventory_contract_version = 'async-adoption-zero-inventory.v1'),
    inventory_sha256 text NOT NULL CHECK (inventory_sha256 ~ '^[0-9a-f]{64}$'),
    queue_count bigint NOT NULL CHECK (queue_count = 0),
    timer_count bigint NOT NULL CHECK (timer_count = 0),
    future_count bigint NOT NULL CHECK (future_count = 0),
    running_count bigint NOT NULL CHECK (running_count = 0),
    observed_at timestamptz NOT NULL,
    receipt_sha256 text NOT NULL CHECK (receipt_sha256 ~ '^[0-9a-f]{64}$'),
    recorded_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    CONSTRAINT async_v1_zero_inventory_time_check CHECK (observed_at <= recorded_at)
);
