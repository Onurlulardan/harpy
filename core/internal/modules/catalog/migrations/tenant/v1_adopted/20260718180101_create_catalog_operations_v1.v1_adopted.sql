-- +goose Up
CREATE TABLE catalog_operation_runs (
    operation_id uuid PRIMARY KEY,
    installation_id uuid NOT NULL,
    package_id uuid NOT NULL,
    kind text NOT NULL CHECK (kind IN ('INSTALL','UPGRADE','RETRY','RESTORE','RETIRE','PURGE')),
    state text NOT NULL CHECK (state IN (
        'PLANNED','APPLYING','WAITING','VERIFYING','ACTIVATING',
        'CANCELLING','SUCCEEDED','BLOCKED','FAILED','CANCELLED'
    )),
    plan jsonb NOT NULL CHECK (jsonb_typeof(plan) = 'object'),
    plan_sha256 text NOT NULL CHECK (plan_sha256 ~ '^[0-9a-f]{64}$'),
    retry_of_operation_id uuid,
    actor_id uuid NOT NULL,
    idempotency_key text NOT NULL CHECK (length(idempotency_key) BETWEEN 8 AND 128),
    request_sha256 text NOT NULL CHECK (request_sha256 ~ '^[0-9a-f]{64}$'),
    started_at timestamptz NOT NULL,
    destructive_started_at timestamptz,
    failure_code text,
    revision bigint NOT NULL CHECK (revision > 0),
    CONSTRAINT catalog_operation_runs_installation_fk
        FOREIGN KEY (installation_id) REFERENCES catalog_installations(installation_id)
        ON DELETE RESTRICT,
    CONSTRAINT catalog_operation_runs_retry_fk
        FOREIGN KEY (retry_of_operation_id) REFERENCES catalog_operation_runs(operation_id)
        ON DELETE RESTRICT,
    UNIQUE (actor_id, idempotency_key),
    CHECK (date_trunc('microseconds', started_at) = started_at),
    CHECK (destructive_started_at IS NULL
        OR date_trunc('microseconds', destructive_started_at) = destructive_started_at)
);

CREATE UNIQUE INDEX catalog_operation_runs_mutating_idx
ON catalog_operation_runs (package_id)
WHERE state IN ('PLANNED','APPLYING','WAITING','VERIFYING','ACTIVATING','CANCELLING');

CREATE TABLE catalog_operation_steps (
    receipt_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL,
    component_id uuid NOT NULL,
    owner_module text NOT NULL,
    owner_generation bigint NOT NULL CHECK (owner_generation > 0),
    owner_contract text NOT NULL,
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    action text NOT NULL CHECK (action IN ('stage','cleanup','activate','disable','enable','retire','restore','purge')),
    ready boolean NOT NULL,
    completed_at timestamptz NOT NULL,
    receipt_sha256 text NOT NULL CHECK (receipt_sha256 ~ '^[0-9a-f]{64}$'),
    UNIQUE (operation_id, component_id, action),
    CONSTRAINT catalog_operation_steps_operation_fk
        FOREIGN KEY (operation_id) REFERENCES catalog_operation_runs(operation_id)
        ON DELETE RESTRICT,
    CHECK (date_trunc('microseconds', completed_at) = completed_at)
);

CREATE TABLE catalog_command_receipts (
    receipt_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL,
    package_id uuid NOT NULL,
    component_id uuid NOT NULL,
    owner_module text NOT NULL,
    owner_generation bigint NOT NULL CHECK (owner_generation > 0),
    owner_contract text NOT NULL,
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    action text NOT NULL CHECK (action IN ('disable','enable')),
    idempotency_key text NOT NULL,
    completed_at timestamptz NOT NULL,
    UNIQUE (operation_id, component_id, action),
    UNIQUE (package_id, component_id, action, idempotency_key),
    CHECK (date_trunc('microseconds', completed_at) = completed_at)
);

CREATE TABLE catalog_purge_blocker_receipts (
    receipt_id uuid PRIMARY KEY,
    operation_id uuid NOT NULL,
    component_id uuid NOT NULL,
    reference_epoch bigint NOT NULL CHECK (reference_epoch > 0),
    catalog_version text NOT NULL,
    zero_digest text NOT NULL CHECK (zero_digest ~ '^[0-9a-f]{64}$'),
    completed_at timestamptz NOT NULL,
    UNIQUE (operation_id, component_id, reference_epoch),
    CONSTRAINT catalog_purge_blockers_operation_fk
        FOREIGN KEY (operation_id) REFERENCES catalog_operation_runs(operation_id)
        ON DELETE RESTRICT,
    CHECK (date_trunc('microseconds', completed_at) = completed_at)
);

CREATE TABLE catalog_demo_receipts (
    receipt_id uuid PRIMARY KEY,
    installation_id uuid NOT NULL,
    operation_id uuid NOT NULL,
    component_id uuid NOT NULL,
    action text NOT NULL CHECK (action IN ('INSTALL','REMOVE')),
    stable_ids jsonb NOT NULL CHECK (jsonb_typeof(stable_ids) = 'array'),
    content_sha256 text NOT NULL CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    actor_id uuid NOT NULL,
    idempotency_key text NOT NULL CHECK (length(idempotency_key) BETWEEN 8 AND 128),
    completed_at timestamptz NOT NULL,
    UNIQUE (installation_id, component_id, action, idempotency_key),
    CONSTRAINT catalog_demo_receipts_installation_fk
        FOREIGN KEY (installation_id) REFERENCES catalog_installations(installation_id)
        ON DELETE RESTRICT,
    CHECK (jsonb_array_length(stable_ids) BETWEEN 1 AND 10000),
    CHECK (date_trunc('microseconds', completed_at) = completed_at)
);
