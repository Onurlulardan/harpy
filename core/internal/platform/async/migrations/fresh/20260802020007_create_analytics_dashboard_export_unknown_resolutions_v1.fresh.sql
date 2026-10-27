-- +goose Up
CREATE TABLE platform.analytics_dashboard_export_unknown_resolutions (
    resolution_receipt_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    export_id uuid NOT NULL,
    job_id uuid NOT NULL,
    attempt_id uuid NOT NULL,
    attempt_generation bigint NOT NULL,
    lease_owner_id uuid NOT NULL,
    effect_id uuid NOT NULL,
    request_sha256 text NOT NULL,
    resolution_kind text NOT NULL,
    resolution_sha256 text NOT NULL,
    resolved_at timestamptz NOT NULL,
    CONSTRAINT analytics_dashboard_export_unknown_resolution_tenant_fk
        FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id)
        ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_export_unknown_resolution_job_fk
        FOREIGN KEY (job_id) REFERENCES platform.jobs(id) ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_export_unknown_resolution_attempt_fk
        FOREIGN KEY (attempt_id) REFERENCES platform.job_attempts(attempt_id)
        ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_export_unknown_resolution_job_unique
        UNIQUE (job_id),
    CONSTRAINT analytics_dashboard_export_unknown_resolution_export_unique
        UNIQUE (tenant_id,export_id),
    CONSTRAINT analytics_dashboard_export_unknown_resolution_attempt_unique
        UNIQUE (attempt_id),
    CONSTRAINT analytics_dashboard_export_unknown_resolution_effect_unique
        UNIQUE (effect_id),
    CONSTRAINT analytics_dashboard_export_unknown_resolution_shape_check CHECK (
        pg_catalog.uuid_extract_version(resolution_receipt_id)=7 AND
        pg_catalog.uuid_extract_version(export_id)=7 AND
        pg_catalog.uuid_extract_version(job_id)=7 AND
        pg_catalog.uuid_extract_version(attempt_id)=7 AND
        attempt_generation BETWEEN 1 AND 3 AND
        pg_catalog.uuid_extract_version(lease_owner_id)=7 AND
        pg_catalog.uuid_extract_version(effect_id)=7 AND
        request_sha256~'^[0-9a-f]{64}$' AND
        request_sha256<>pg_catalog.repeat('0',64) AND
        resolution_kind IN ('deleted','absent') AND
        resolution_sha256~'^[0-9a-f]{64}$' AND
        resolution_sha256<>pg_catalog.repeat('0',64) AND
        pg_catalog.isfinite(resolved_at)
    )
);

CREATE INDEX analytics_dashboard_export_unknown_resolution_lookup_idx
ON platform.analytics_dashboard_export_unknown_resolutions(
    tenant_id,export_id,attempt_generation,lease_owner_id
);

CREATE TRIGGER analytics_dashboard_export_unknown_resolution_append_only_guard
BEFORE UPDATE OR DELETE
ON platform.analytics_dashboard_export_unknown_resolutions
FOR EACH ROW EXECUTE FUNCTION platform.platform_async_guard_append_only_v1();

ALTER TABLE platform.analytics_dashboard_export_unknown_resolutions
ENABLE ALWAYS TRIGGER analytics_dashboard_export_unknown_resolution_append_only_guard;

REVOKE ALL ON TABLE
    platform.analytics_dashboard_export_unknown_resolutions FROM PUBLIC;
