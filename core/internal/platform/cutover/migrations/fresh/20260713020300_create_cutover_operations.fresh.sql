-- +goose Up
CREATE TABLE platform.cutover_operations (
    operation_id uuid NOT NULL,
    run_id uuid NOT NULL,
    lease_id uuid NOT NULL,
    release_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    group_id text NOT NULL,
    domain text NOT NULL,
    cutover_generation bigint NOT NULL,
    lease_generation bigint NOT NULL,
    lease_owner uuid NOT NULL,
    lease_expires_at timestamptz NOT NULL,
    last_heartbeat_at timestamptz NOT NULL,
    lease_manifest_sha256 text NOT NULL,
    plan_sha256 text NOT NULL,
    transform_manifest_sha256 text NOT NULL,
    source_system_identifier numeric(20, 0) NOT NULL,
    source_database_oid bigint NOT NULL,
    target_system_identifier numeric(20, 0) NOT NULL,
    target_database_oid bigint NOT NULL,
    publication_name text NOT NULL,
    slot_name text NOT NULL,
    state text NOT NULL,
    blocked_reason text,
    created_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    PRIMARY KEY (operation_id, run_id),
    CONSTRAINT cutover_operations_lease_unique UNIQUE (lease_id),
    CONSTRAINT cutover_operations_domain_fk
        FOREIGN KEY (scope_kind, scope_id, domain)
        REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_operations_group_fk
        FOREIGN KEY (scope_kind, scope_id, group_id)
        REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON DELETE RESTRICT,
    CONSTRAINT cutover_operations_group_member_fk
        FOREIGN KEY (scope_kind, scope_id, group_id, domain)
        REFERENCES platform.cutover_group_members (
            scope_kind, scope_id, group_id, domain
        )
        ON DELETE RESTRICT,
    CONSTRAINT cutover_operations_scope_check
        CHECK (scope_kind IN ('global', 'tenant')),
    CONSTRAINT cutover_operations_generation_check
        CHECK (cutover_generation > 0 AND lease_generation > 0),
    CONSTRAINT cutover_operations_digest_check
        CHECK (
            lease_manifest_sha256 ~ '^[0-9a-f]{64}$' AND
            plan_sha256 ~ '^[0-9a-f]{64}$' AND
            transform_manifest_sha256 ~ '^[0-9a-f]{64}$'
        ),
    CONSTRAINT cutover_operations_database_check
        CHECK (
            source_system_identifier > 0 AND
            source_system_identifier = target_system_identifier AND
            source_database_oid BETWEEN 1 AND 4294967295 AND
            source_database_oid = target_database_oid
        ),
    CONSTRAINT cutover_operations_publication_check
        CHECK (publication_name ~ '^[a-z][a-z0-9_]{0,62}$'),
    CONSTRAINT cutover_operations_slot_check
        CHECK (slot_name ~ '^[a-z][a-z0-9_]{0,62}$' AND slot_name <> publication_name),
    CONSTRAINT cutover_operations_state_check
        CHECK (state IN (
            'planned', 'copying', 'shadow', 'quiescing', 'blocked',
            'completed', 'failed', 'cancelled'
        )),
    CONSTRAINT cutover_operations_blocked_reason_check
        CHECK (
            (state = 'blocked' AND pg_catalog.octet_length(blocked_reason) BETWEEN 1 AND 2048) OR
            (state <> 'blocked' AND blocked_reason IS NULL)
        ),
    CONSTRAINT cutover_operations_time_check
        CHECK (last_heartbeat_at <= updated_at AND created_at <= updated_at)
);

CREATE INDEX cutover_operations_state_idx
    ON platform.cutover_operations (state, lease_expires_at, last_heartbeat_at);

CREATE INDEX cutover_operations_scope_idx
    ON platform.cutover_operations (scope_kind, scope_id, group_id, domain, cutover_generation);
