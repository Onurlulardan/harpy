-- +goose Up
CREATE TABLE platform.jobs (
    id uuid PRIMARY KEY,
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
    kind text NOT NULL CHECK (kind ~ '^[a-z][a-z0-9_.-]{2,127}$'),
    payload_revision integer NOT NULL CHECK (payload_revision > 0),
    payload_ref text,
    payload_sha256 text NOT NULL CHECK (payload_sha256 ~ '^[0-9a-f]{64}$'),
    payload_redacted_at timestamptz,
    dedupe_key text NOT NULL
        CHECK (pg_catalog.octet_length(dedupe_key) BETWEEN 1 AND 512),
    effect_class text NOT NULL
        CHECK (effect_class IN ('database_transactional', 'external')),
    effect_idempotency_key text
        CHECK (pg_catalog.octet_length(effect_idempotency_key) BETWEEN 1 AND 512),
    priority integer NOT NULL,
    run_at timestamptz NOT NULL,
    state text NOT NULL CHECK (state IN (
        'queued', 'running', 'succeeded', 'dead', 'cancelled'
    )),
    attempts integer NOT NULL CHECK (attempts >= 0),
    max_attempts integer NOT NULL CHECK (max_attempts BETWEEN 1 AND 25),
    lease_owner_id uuid,
    lease_generation bigint NOT NULL CHECK (lease_generation >= 0),
    lease_heartbeat_at timestamptz,
    lease_until timestamptz,
    cancel_requested_at timestamptz,
    parked_at timestamptz,
    parked_code text CHECK (parked_code ~ '^[A-Z][A-Z0-9_]{1,63}$'),
    parked_current_group_generation bigint
        CHECK (parked_current_group_generation > 0),
    parked_current_cutover_generation bigint
        CHECK (parked_current_cutover_generation > 0),
    effect_started_at timestamptz,
    last_error_class text CHECK (last_error_class IN (
        'transient', 'permanent', 'conflict', 'cancelled'
    )),
    last_error_code text CHECK (last_error_code ~ '^[A-Z][A-Z0-9_]{1,63}$'),
    last_error_sha256 text CHECK (last_error_sha256 ~ '^[0-9a-f]{64}$'),
    retry_of_job_id uuid,
    retain_until timestamptz,
    retention_release_revision bigint NOT NULL DEFAULT 0
        CHECK (retention_release_revision >= 0),
    retention_release_sha256 text
        CHECK (retention_release_sha256 ~ '^[0-9a-f]{64}$'),
    retention_released_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    updated_at timestamptz NOT NULL DEFAULT pg_catalog.clock_timestamp(),
    completed_at timestamptz,
    CONSTRAINT jobs_tenant_fk FOREIGN KEY (tenant_id)
        REFERENCES platform.tenants (id) ON DELETE RESTRICT,
    CONSTRAINT jobs_authority_fk FOREIGN KEY (scope_kind, scope_id, authority_domain)
        REFERENCES platform.domain_cutovers (scope_kind, scope_id, domain)
        ON DELETE RESTRICT,
    CONSTRAINT jobs_authority_group_fk
        FOREIGN KEY (scope_kind, scope_id, authority_group_id)
        REFERENCES platform.cutover_groups (scope_kind, scope_id, group_id)
        ON DELETE RESTRICT,
    CONSTRAINT jobs_authority_member_fk
        FOREIGN KEY (scope_kind, scope_id, authority_group_id, authority_domain)
        REFERENCES platform.cutover_group_members (
            scope_kind, scope_id, group_id, domain
        ) ON DELETE RESTRICT,
    CONSTRAINT jobs_retry_fk FOREIGN KEY (retry_of_job_id)
        REFERENCES platform.jobs (id) ON DELETE RESTRICT,
    CONSTRAINT jobs_scope_identity_unique UNIQUE (scope_kind, scope_id, id),
    CONSTRAINT jobs_dedupe_unique UNIQUE (
        scope_kind, scope_id, authority_group_id, authority_domain, kind, dedupe_key
    ),
    CONSTRAINT jobs_scope_check CHECK (
        (scope_kind = 'tenant' AND tenant_id IS NOT NULL AND tenant_id = scope_id) OR
        (scope_kind = 'global' AND tenant_id IS NULL)
    ),
    CONSTRAINT jobs_payload_check CHECK (
        (payload_ref IS NOT NULL AND
            pg_catalog.octet_length(payload_ref) BETWEEN 1 AND 2048 AND
            payload_redacted_at IS NULL) OR
        (payload_ref IS NULL AND payload_redacted_at IS NOT NULL)
    ),
    CONSTRAINT jobs_effect_check CHECK (
        effect_class = 'database_transactional' OR effect_idempotency_key IS NOT NULL
    ),
    CONSTRAINT jobs_attempt_check CHECK (
        attempts <= max_attempts AND lease_generation = attempts
    ),
    CONSTRAINT jobs_lease_check CHECK (
        (state = 'running' AND attempts > 0 AND lease_owner_id IS NOT NULL AND
            lease_heartbeat_at IS NOT NULL AND lease_until IS NOT NULL AND
            lease_heartbeat_at < lease_until) OR
        (state <> 'running' AND lease_owner_id IS NULL AND
            lease_heartbeat_at IS NULL AND lease_until IS NULL)
    ),
    CONSTRAINT jobs_parked_check CHECK (
        (parked_at IS NULL AND parked_code IS NULL AND
            parked_current_group_generation IS NULL AND
            parked_current_cutover_generation IS NULL) OR
        (state IN ('queued', 'cancelled') AND
            parked_at IS NOT NULL AND parked_code IS NOT NULL AND
            parked_current_group_generation >= authority_group_generation AND
            parked_current_cutover_generation >= cutover_generation AND
            (parked_current_group_generation > authority_group_generation OR
             parked_current_cutover_generation > cutover_generation))
    ),
    CONSTRAINT jobs_outcome_check CHECK (
        (state IN ('queued', 'running') AND completed_at IS NULL) OR
        (state IN ('succeeded', 'dead', 'cancelled') AND completed_at IS NOT NULL)
    ),
    CONSTRAINT jobs_retention_check CHECK (
        (state IN ('queued', 'running') AND retain_until IS NULL AND
            retention_release_revision = 0 AND
            retention_release_sha256 IS NULL AND retention_released_at IS NULL) OR
        (state IN ('succeeded', 'cancelled') AND
            retain_until >= completed_at + interval '90 days') OR
        (state = 'dead' AND retain_until >= completed_at + interval '365 days')
    ),
    CONSTRAINT jobs_retention_release_check CHECK (
        (retention_release_revision = 0 AND
            retention_release_sha256 IS NULL AND retention_released_at IS NULL) OR
        (retention_release_revision > 0 AND
            retention_release_sha256 IS NOT NULL AND retention_released_at IS NOT NULL)
    ),
    CONSTRAINT jobs_error_check CHECK (
        (state IN ('queued', 'running', 'succeeded') AND
            last_error_class IS NULL AND last_error_code IS NULL AND
            last_error_sha256 IS NULL) OR
        (state = 'dead' AND last_error_class IS NULL AND
            last_error_code = 'JOB_OUTCOME_UNKNOWN' AND
            last_error_sha256 IS NULL AND effect_started_at IS NOT NULL) OR
        (state IN ('queued', 'running', 'dead', 'cancelled') AND
            last_error_class IS NOT NULL AND last_error_code IS NOT NULL AND
            last_error_sha256 IS NOT NULL)
    ),
    CONSTRAINT jobs_time_check CHECK (
        created_at <= updated_at AND
        (payload_redacted_at IS NULL OR completed_at <= payload_redacted_at) AND
        (payload_redacted_at IS NULL OR state IN ('succeeded', 'dead', 'cancelled')) AND
        (cancel_requested_at IS NULL OR created_at <= cancel_requested_at) AND
        (parked_at IS NULL OR created_at <= parked_at) AND
        (effect_started_at IS NULL OR created_at <= effect_started_at) AND
        (completed_at IS NULL OR updated_at <= completed_at) AND
        (retention_released_at IS NULL OR completed_at <= retention_released_at)
    )
);
