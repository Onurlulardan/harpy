-- +goose Up
ALTER TABLE analytics_saved_queries
    ALTER COLUMN current_draft_version_id DROP NOT NULL;
ALTER TABLE analytics_saved_queries
    ADD CONSTRAINT analytics_saved_queries_draft_presence_check CHECK (
        (lifecycle_state = 'active' AND current_draft_version_id IS NOT NULL)
        OR lifecycle_state = 'retired'
    );
ALTER TABLE analytics_dashboards
    ALTER COLUMN current_draft_version_id DROP NOT NULL;
ALTER TABLE analytics_dashboards
    ADD CONSTRAINT analytics_dashboards_draft_presence_check CHECK (
        (lifecycle_state = 'active' AND current_draft_version_id IS NOT NULL)
        OR lifecycle_state = 'retired'
    );
ALTER TABLE analytics_publication_approvals
    DROP CONSTRAINT analytics_publication_approvals_release_fk;

CREATE TABLE analytics_inbound_references (
    reference_id uuid NOT NULL PRIMARY KEY,
    target_kind text NOT NULL,
    target_id uuid NOT NULL,
    source_module text NOT NULL,
    source_kind text NOT NULL,
    source_id uuid NOT NULL,
    source_version_id uuid NOT NULL,
    usage text NOT NULL,
    state text NOT NULL,
    created_at timestamptz NOT NULL,
    released_by uuid,
    released_at timestamptz,
    CONSTRAINT analytics_inbound_references_identity_key UNIQUE (
        target_kind,target_id,source_module,source_kind,
        source_id,source_version_id,usage
    ),
    CONSTRAINT analytics_inbound_references_shape_check CHECK (
        pg_catalog.uuid_extract_version(reference_id) = 7
        AND target_kind IN ('query_version','dashboard_version','publication_release')
        AND pg_catalog.uuid_extract_version(target_id) = 7
        AND source_module ~ '^[a-z][a-z0-9_]{0,63}$'
        AND source_kind ~ '^[a-z][a-z0-9_.-]{0,63}$'
        AND pg_catalog.uuid_extract_version(source_id) = 7
        AND pg_catalog.uuid_extract_version(source_version_id) = 7
        AND usage IN ('read','write','schema','lifecycle')
        AND state IN ('active','released')
        AND pg_catalog.isfinite(created_at)
        AND ((state = 'active' AND released_by IS NULL AND released_at IS NULL)
          OR (state = 'released'
              AND released_by <> '00000000-0000-0000-0000-000000000000'::uuid
              AND pg_catalog.isfinite(released_at)
              AND released_at >= created_at))
    )
);

CREATE TABLE analytics_legal_holds (
    hold_id uuid NOT NULL PRIMARY KEY,
    reason_sha256 bytea NOT NULL,
    placed_by uuid NOT NULL,
    placed_at timestamptz NOT NULL,
    state text NOT NULL,
    row_revision bigint NOT NULL,
    released_by uuid,
    released_at timestamptz,
    CONSTRAINT analytics_legal_holds_shape_check CHECK (
        pg_catalog.uuid_extract_version(hold_id) = 7
        AND pg_catalog.octet_length(reason_sha256) = 32
        AND placed_by <> '00000000-0000-0000-0000-000000000000'::uuid
        AND pg_catalog.isfinite(placed_at)
        AND state IN ('active','released') AND row_revision > 0
        AND ((state = 'active' AND released_by IS NULL AND released_at IS NULL)
          OR (state = 'released'
              AND released_by <> '00000000-0000-0000-0000-000000000000'::uuid
              AND pg_catalog.isfinite(released_at) AND released_at >= placed_at))
    )
);

CREATE TABLE analytics_retention_checkpoints (
    singleton boolean NOT NULL PRIMARY KEY DEFAULT true CHECK (singleton),
    revision bigint NOT NULL CHECK (revision >= 0),
    cycle_count bigint NOT NULL CHECK (cycle_count >= 0),
    last_job_id uuid,
    last_attempt_id uuid,
    last_completed_at timestamptz,
    CHECK ((revision = 0) = (last_job_id IS NULL)),
    CHECK ((last_job_id IS NULL) = (last_attempt_id IS NULL)),
    CHECK ((last_job_id IS NULL) = (last_completed_at IS NULL))
);
INSERT INTO analytics_retention_checkpoints (
    singleton,revision,cycle_count,last_job_id,last_attempt_id,last_completed_at
) VALUES (true,0,0,NULL,NULL,NULL);

CREATE TABLE analytics_retention_receipts (
    job_id uuid NOT NULL PRIMARY KEY,
    attempt_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    payload_reference text COLLATE "C" NOT NULL,
    payload_sha256 bytea NOT NULL,
    successor_job_id uuid NOT NULL UNIQUE,
    successor_run_at timestamptz NOT NULL,
    successor_dedupe_key text COLLATE "C" NOT NULL,
    drafts_purged bigint NOT NULL,
    versions_purged bigint NOT NULL,
    releases_purged bigint NOT NULL,
    executions_purged bigint NOT NULL,
    audits_purged bigint NOT NULL,
    capabilities_purged bigint NOT NULL,
    legal_hold_blocked boolean NOT NULL,
    has_remaining boolean NOT NULL,
    checkpoint_revision bigint NOT NULL,
    receipt_sha256 bytea NOT NULL,
    completed_at timestamptz NOT NULL,
    CONSTRAINT analytics_retention_receipts_shape_check CHECK (
        pg_catalog.uuid_extract_version(job_id) = 7
        AND pg_catalog.uuid_extract_version(attempt_id) = 7
        AND tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid
        AND pg_catalog.octet_length(payload_reference) BETWEEN 1 AND 2048
        AND pg_catalog.octet_length(payload_sha256) = 32
        AND pg_catalog.uuid_extract_version(successor_job_id) = 7
        AND pg_catalog.isfinite(successor_run_at)
        AND pg_catalog.octet_length(successor_dedupe_key) BETWEEN 1 AND 256
        AND drafts_purged >= 0 AND versions_purged >= 0
        AND releases_purged >= 0 AND executions_purged >= 0
        AND audits_purged >= 0 AND capabilities_purged >= 0
        AND checkpoint_revision > 0
        AND pg_catalog.octet_length(receipt_sha256) = 32
        AND pg_catalog.isfinite(completed_at)
        AND successor_run_at > completed_at
    )
);

CREATE TABLE analytics_retention_contexts (
    backend_pid integer NOT NULL,
    transaction_id xid8 NOT NULL,
    job_id uuid NOT NULL,
    opened_at timestamptz NOT NULL,
    PRIMARY KEY (backend_pid,transaction_id),
    CHECK (backend_pid > 0 AND pg_catalog.uuid_extract_version(job_id) = 7
           AND pg_catalog.isfinite(opened_at))
);

CREATE INDEX analytics_inbound_references_blocker_idx
    ON analytics_inbound_references (target_kind,target_id)
    WHERE state = 'active';
CREATE UNIQUE INDEX analytics_legal_holds_single_active_idx
    ON analytics_legal_holds ((state)) WHERE state = 'active';
CREATE INDEX analytics_query_versions_retention_idx
    ON analytics_saved_query_versions (updated_at,id) WHERE version_state = 'retired';
CREATE INDEX analytics_dashboard_versions_retention_idx
    ON analytics_dashboard_versions (updated_at,id) WHERE version_state = 'retired';
CREATE INDEX analytics_publication_releases_retention_idx
    ON analytics_publication_releases (revoked_at,id)
    WHERE lifecycle_state IN ('revoked','retired');
CREATE INDEX analytics_capability_tokens_retention_idx
    ON analytics_capability_tokens (expires_at,release_id);
