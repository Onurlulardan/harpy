-- +goose Up
CREATE TABLE platform.analytics_dashboard_export_maintenance_unknown_claims (
    job_id uuid PRIMARY KEY,
    kind text NOT NULL,
    tenant_id uuid NOT NULL,
    tenant_schema name NOT NULL,
    export_id uuid NOT NULL,
    authority_group_id text NOT NULL,
    authority_group_generation bigint NOT NULL,
    cutover_generation bigint NOT NULL,
    attempt_id uuid NOT NULL,
    attempt_number integer NOT NULL,
    lease_generation bigint NOT NULL,
    original_lease_owner_id uuid NOT NULL,
    effect_id uuid NOT NULL,
    request_sha256 text NOT NULL,
    unknown_at timestamptz NOT NULL,
    claim_generation bigint NOT NULL DEFAULT 0,
    cleanup_lease_owner_id uuid,
    lease_heartbeat_at timestamptz,
    lease_until timestamptz,
    run_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT analytics_dashboard_export_maintenance_claim_tenant_fk
        FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_export_maintenance_claim_job_fk
        FOREIGN KEY (job_id) REFERENCES platform.jobs(id) ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_export_maintenance_claim_attempt_fk
        FOREIGN KEY (attempt_id) REFERENCES platform.job_attempts(attempt_id)
        ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_export_maintenance_claim_attempt_unique
        UNIQUE (attempt_id),
    CONSTRAINT analytics_dashboard_export_maintenance_claim_effect_unique
        UNIQUE (effect_id),
    CONSTRAINT analytics_dashboard_export_maintenance_claim_export_unique
        UNIQUE (tenant_id,kind,export_id),
    CONSTRAINT analytics_dashboard_export_maintenance_claim_shape_check CHECK (
        kind IN ('analytics.dashboard.export.delete',
                 'analytics.dashboard.export.expire') AND
        tenant_schema::text~'^[a-z][a-z0-9_]{0,62}$' AND
        pg_catalog.uuid_extract_version(export_id)=7 AND
        pg_catalog.uuid_extract_version(job_id)=7 AND
        authority_group_id~'^[a-z][a-z0-9_]{0,62}$' AND
        authority_group_generation>0 AND cutover_generation>0 AND
        pg_catalog.uuid_extract_version(attempt_id)=7 AND
        attempt_number BETWEEN 1 AND
            CASE kind WHEN 'analytics.dashboard.export.delete' THEN 3 ELSE 8 END AND
        lease_generation>0 AND
        pg_catalog.uuid_extract_version(original_lease_owner_id)=7 AND
        pg_catalog.uuid_extract_version(effect_id)=5 AND
        request_sha256~'^[0-9a-f]{64}$' AND request_sha256<>repeat('0',64) AND
        pg_catalog.isfinite(unknown_at)
    ),
    CONSTRAINT analytics_dashboard_export_maintenance_claim_lease_check CHECK (
        claim_generation>=0 AND (
            cleanup_lease_owner_id IS NULL AND lease_heartbeat_at IS NULL AND
            lease_until IS NULL OR
            pg_catalog.uuid_extract_version(cleanup_lease_owner_id)=7 AND
            claim_generation>0 AND lease_heartbeat_at IS NOT NULL AND
            lease_until IS NOT NULL AND lease_heartbeat_at<lease_until
        )
    ),
    CONSTRAINT analytics_dashboard_export_maintenance_claim_time_check CHECK (
        pg_catalog.isfinite(run_at) AND pg_catalog.isfinite(created_at) AND
        pg_catalog.isfinite(updated_at) AND created_at<=updated_at
    )
);

CREATE INDEX analytics_dashboard_export_maintenance_unknown_due_idx
ON platform.analytics_dashboard_export_maintenance_unknown_claims(
    run_at,lease_until,job_id
);

CREATE TABLE platform.analytics_dashboard_export_maintenance_unknown_resolutions (
    resolution_receipt_id uuid PRIMARY KEY,
    effect_receipt_id uuid NOT NULL UNIQUE,
    effect_receipt_sha256 text NOT NULL,
    resolution_sha256 text NOT NULL,
    kind text NOT NULL,
    tenant_id uuid NOT NULL,
    tenant_schema name NOT NULL,
    export_id uuid NOT NULL,
    authority_group_id text NOT NULL,
    authority_group_generation bigint NOT NULL,
    cutover_generation bigint NOT NULL,
    job_id uuid NOT NULL UNIQUE,
    attempt_id uuid NOT NULL UNIQUE,
    attempt_number integer NOT NULL,
    lease_generation bigint NOT NULL,
    original_lease_owner_id uuid NOT NULL,
    effect_id uuid NOT NULL UNIQUE,
    request_sha256 text NOT NULL,
    original_unknown_at timestamptz NOT NULL,
    original_attempt_finished_at timestamptz NOT NULL,
    original_attempt_outcome text NOT NULL,
    original_error_class text,
    original_error_code text NOT NULL,
    original_error_sha256 text,
    cleanup_generation bigint NOT NULL,
    cleanup_lease_owner_id uuid NOT NULL,
    terminal_state text NOT NULL,
    provider_code text NOT NULL,
    provider_reference text NOT NULL,
    result_reference text NOT NULL,
    result_sha256 text NOT NULL,
    resolved_at timestamptz NOT NULL,
    FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE RESTRICT,
    FOREIGN KEY (job_id) REFERENCES platform.jobs(id) ON DELETE RESTRICT,
    FOREIGN KEY (attempt_id) REFERENCES platform.job_attempts(attempt_id)
        ON DELETE RESTRICT,
    FOREIGN KEY (effect_receipt_id)
        REFERENCES platform.idempotency_receipts(receipt_id) ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_export_maintenance_resolution_shape_check CHECK (
        pg_catalog.uuid_extract_version(resolution_receipt_id)=7 AND
        pg_catalog.uuid_extract_version(effect_receipt_id)=5 AND
        effect_receipt_sha256~'^[0-9a-f]{64}$' AND
        effect_receipt_sha256<>repeat('0',64) AND
        resolution_sha256~'^[0-9a-f]{64}$' AND resolution_sha256<>repeat('0',64) AND
        kind IN ('analytics.dashboard.export.delete',
                 'analytics.dashboard.export.expire') AND
        tenant_schema::text~'^[a-z][a-z0-9_]{0,62}$' AND
        pg_catalog.uuid_extract_version(export_id)=7 AND
        pg_catalog.uuid_extract_version(job_id)=7 AND
        authority_group_id~'^[a-z][a-z0-9_]{0,62}$' AND
        authority_group_generation>0 AND cutover_generation>0 AND
        pg_catalog.uuid_extract_version(attempt_id)=7 AND
        attempt_number BETWEEN 1 AND
            CASE kind WHEN 'analytics.dashboard.export.delete' THEN 3 ELSE 8 END AND
        lease_generation>0 AND
        pg_catalog.uuid_extract_version(original_lease_owner_id)=7 AND
        pg_catalog.uuid_extract_version(effect_id)=5 AND
        request_sha256~'^[0-9a-f]{64}$' AND request_sha256<>repeat('0',64) AND
        original_attempt_outcome='dead' AND original_error_class IS NULL AND
        original_error_code='JOB_OUTCOME_UNKNOWN' AND original_error_sha256 IS NULL AND
        cleanup_generation>0 AND
        pg_catalog.uuid_extract_version(cleanup_lease_owner_id)=7 AND
        ((kind='analytics.dashboard.export.delete' AND terminal_state='deleted' AND
          provider_code='ANALYTICS_EXPORT_DELETED') OR
         (kind='analytics.dashboard.export.expire' AND
          terminal_state IN ('expired','cancelled','failed','deleted') AND
          provider_code='ANALYTICS_EXPORT_EXPIRY_COMPLETE')) AND
        provider_reference=export_id::text AND
        result_sha256~'^[0-9a-f]{64}$' AND result_sha256<>repeat('0',64) AND
        pg_catalog.isfinite(original_unknown_at) AND
        original_attempt_finished_at=original_unknown_at AND
        pg_catalog.isfinite(resolved_at) AND resolved_at>original_unknown_at
    )
);

CREATE TRIGGER analytics_dashboard_export_maintenance_resolution_append_guard
BEFORE UPDATE OR DELETE
ON platform.analytics_dashboard_export_maintenance_unknown_resolutions
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_guard_append_only_v1();
ALTER TABLE platform.analytics_dashboard_export_maintenance_unknown_resolutions
ENABLE ALWAYS TRIGGER analytics_dashboard_export_maintenance_resolution_append_guard;
REVOKE ALL ON TABLE
    platform.analytics_dashboard_export_maintenance_unknown_claims,
    platform.analytics_dashboard_export_maintenance_unknown_resolutions FROM PUBLIC;
