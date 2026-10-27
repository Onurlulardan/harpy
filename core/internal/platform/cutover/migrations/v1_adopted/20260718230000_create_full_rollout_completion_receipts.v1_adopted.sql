-- +goose Up
CREATE TABLE cutover_authority.full_rollout_completion_receipts (
    receipt_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    evidence_sha256 bytea NOT NULL,
    shadow_read_sha256 bytea NOT NULL,
    logical_delta_sha256 bytea NOT NULL,
    reconcile_sha256 bytea NOT NULL,
    production_snapshot_sha256 bytea NOT NULL,
    v1_removal_sha256 bytea NOT NULL,
    backup_restore_sha256 bytea NOT NULL,
    frontend_contract_sha256 bytea NOT NULL,
    legacy_route_hits bigint NOT NULL,
    v1_writers bigint NOT NULL,
    v1_consumers bigint NOT NULL,
    v1_queue_messages bigint NOT NULL,
    v1_credentials bigint NOT NULL,
    last_legacy_hit_at timestamptz NOT NULL,
    rollback_window_closed_at timestamptz NOT NULL,
    recorded_at timestamptz NOT NULL,
    recorded_by name NOT NULL,
    CONSTRAINT full_rollout_completion_scope_key
        UNIQUE (environment_id, tenant_id),
    CONSTRAINT full_rollout_completion_identity_check CHECK (
        environment_id <> '00000000-0000-0000-0000-000000000000'::uuid AND
        tenant_id <> '00000000-0000-0000-0000-000000000000'::uuid AND
        environment_id <> tenant_id
    ),
    CONSTRAINT full_rollout_completion_digest_check CHECK (
        pg_catalog.octet_length(evidence_sha256) = 32 AND
        pg_catalog.octet_length(shadow_read_sha256) = 32 AND
        pg_catalog.octet_length(logical_delta_sha256) = 32 AND
        pg_catalog.octet_length(reconcile_sha256) = 32 AND
        pg_catalog.octet_length(production_snapshot_sha256) = 32 AND
        pg_catalog.octet_length(v1_removal_sha256) = 32 AND
        pg_catalog.octet_length(backup_restore_sha256) = 32 AND
        pg_catalog.octet_length(frontend_contract_sha256) = 32
    ),
    CONSTRAINT full_rollout_completion_zero_v1_check CHECK (
        legacy_route_hits = 0 AND v1_writers = 0 AND v1_consumers = 0 AND
        v1_queue_messages = 0 AND v1_credentials = 0
    ),
    CONSTRAINT full_rollout_completion_time_check CHECK (
        last_legacy_hit_at <= recorded_at - interval '30 days' AND
        rollback_window_closed_at <= recorded_at
    ),
    CONSTRAINT full_rollout_completion_actor_check CHECK (
        recorded_by = 'harpy_cluster_admin'
    )
);

CREATE TRIGGER full_rollout_completion_receipts_immutable
BEFORE UPDATE OR DELETE ON cutover_authority.full_rollout_completion_receipts
FOR EACH STATEMENT EXECUTE FUNCTION platform.cutover_reject_mutation();
ALTER TABLE cutover_authority.full_rollout_completion_receipts
    ENABLE ALWAYS TRIGGER full_rollout_completion_receipts_immutable;

REVOKE ALL ON ALL TABLES IN SCHEMA cutover_authority FROM PUBLIC;
REVOKE ALL ON ALL TABLES IN SCHEMA cutover_authority FROM harpy_migrator;
REVOKE ALL ON ALL TABLES IN SCHEMA cutover_authority FROM harpy_cluster_admin;
REVOKE ALL ON ALL TABLES IN SCHEMA cutover_authority FROM harpy_cutover_finalizer;
REVOKE ALL ON ALL TABLES IN SCHEMA cutover_authority FROM harpy_adopter;
REVOKE ALL ON ALL TABLES IN SCHEMA cutover_authority FROM harpy_runtime;
