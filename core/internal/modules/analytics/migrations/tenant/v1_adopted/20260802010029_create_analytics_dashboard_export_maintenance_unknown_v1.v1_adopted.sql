-- +goose Up
ALTER TABLE analytics_dashboard_exports
    ADD COLUMN maintenance_unknown_job_id uuid,
    ADD COLUMN maintenance_cleanup_generation bigint,
    ADD COLUMN maintenance_cleanup_lease_owner uuid,
    ADD COLUMN maintenance_cleanup_lease_until timestamptz,
    ADD COLUMN maintenance_local_state_sha256 bytea,
    ADD COLUMN maintenance_local_state_at timestamptz,
    ADD CONSTRAINT analytics_dashboard_export_maintenance_cleanup_check CHECK (
        (maintenance_unknown_job_id IS NULL AND
         maintenance_cleanup_generation IS NULL AND
         maintenance_cleanup_lease_owner IS NULL AND
         maintenance_cleanup_lease_until IS NULL AND
         maintenance_local_state_sha256 IS NULL AND
         maintenance_local_state_at IS NULL) OR
        (pg_catalog.uuid_extract_version(maintenance_unknown_job_id)=7 AND
         maintenance_cleanup_generation>=1 AND (
            maintenance_cleanup_lease_owner IS NULL AND
            maintenance_cleanup_lease_until IS NULL OR
            pg_catalog.uuid_extract_version(maintenance_cleanup_lease_owner)=7 AND
            maintenance_cleanup_lease_until IS NOT NULL AND
            pg_catalog.isfinite(maintenance_cleanup_lease_until)
         ) AND pg_catalog.octet_length(maintenance_local_state_sha256)=32 AND
         pg_catalog.encode(maintenance_local_state_sha256,'hex') IS DISTINCT FROM
            pg_catalog.repeat('0',64)
         AND pg_catalog.isfinite(maintenance_local_state_at))
    );

CREATE TABLE analytics_dashboard_export_maintenance_unknown_resolutions (
    resolution_receipt_id uuid PRIMARY KEY,
    effect_receipt_id uuid NOT NULL UNIQUE,
    effect_receipt_sha256 bytea NOT NULL,
    resolution_sha256 bytea NOT NULL,
    tenant_id uuid NOT NULL,
    tenant_schema name NOT NULL,
    kind text NOT NULL,
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
    effect_request_sha256 bytea NOT NULL,
    original_unknown_at timestamptz NOT NULL,
    cleanup_generation bigint NOT NULL,
    cleanup_lease_owner_id uuid NOT NULL,
    delete_attempt_generation bigint NOT NULL,
    resolution_kind text NOT NULL,
    terminal_state text NOT NULL,
    local_state_sha256 bytea NOT NULL,
    local_state_at timestamptz NOT NULL,
    object_key text,
    object_version text,
    reconcile_provider_request_id text,
    reconcile_evidence_sha256 bytea,
    upload_completed_at timestamptz,
    reconcile_verified_at timestamptz,
    delete_provider_request_id text,
    delete_response_sha256 bytea,
    absence_sha256 bytea,
    delete_verified_at timestamptz,
    quota_released_bytes bigint NOT NULL,
    resolved_at timestamptz NOT NULL,
    CHECK(pg_catalog.uuid_extract_version(resolution_receipt_id)=7
        AND pg_catalog.uuid_extract_version(effect_receipt_id)=5
        AND pg_catalog.uuid_extract_version(export_id)=7
        AND pg_catalog.uuid_extract_version(job_id)=7
        AND pg_catalog.uuid_extract_version(attempt_id)=7
        AND pg_catalog.uuid_extract_version(original_lease_owner_id)=7
        AND pg_catalog.uuid_extract_version(effect_id)=5
        AND pg_catalog.uuid_extract_version(cleanup_lease_owner_id)=7),
    CHECK(kind IN ('analytics.dashboard.export.delete',
                   'analytics.dashboard.export.expire')
        AND tenant_schema::text~'^[a-z][a-z0-9_]{0,62}$'
        AND authority_group_id~'^[a-z][a-z0-9_]{0,62}$'
        AND authority_group_generation>0 AND cutover_generation>0
        AND attempt_number BETWEEN 1 AND
            (CASE kind WHEN 'analytics.dashboard.export.delete' THEN 3 ELSE 8 END)
        AND lease_generation=attempt_number AND cleanup_generation>=1
        AND delete_attempt_generation BETWEEN 0 AND 3),
    CHECK(resolution_kind IN ('deleted','absent')
        AND terminal_state IN ('deleted','expired','cancelled','failed')
        AND (kind IS DISTINCT FROM 'analytics.dashboard.export.delete' OR
             terminal_state='deleted')),
    CHECK(pg_catalog.octet_length(effect_receipt_sha256)=32
        AND pg_catalog.octet_length(resolution_sha256)=32
        AND pg_catalog.octet_length(effect_request_sha256)=32
        AND pg_catalog.octet_length(local_state_sha256)=32
        AND pg_catalog.encode(effect_receipt_sha256,'hex') IS DISTINCT FROM
            pg_catalog.repeat('0',64)
        AND pg_catalog.encode(resolution_sha256,'hex') IS DISTINCT FROM
            pg_catalog.repeat('0',64)
        AND pg_catalog.encode(effect_request_sha256,'hex') IS DISTINCT FROM
            pg_catalog.repeat('0',64)
        AND pg_catalog.encode(local_state_sha256,'hex') IS DISTINCT FROM
            pg_catalog.repeat('0',64)
        AND (reconcile_evidence_sha256 IS NULL OR
             pg_catalog.encode(reconcile_evidence_sha256,'hex') IS DISTINCT FROM
                pg_catalog.repeat('0',64))
        AND (delete_response_sha256 IS NULL OR
             pg_catalog.encode(delete_response_sha256,'hex') IS DISTINCT FROM
                pg_catalog.repeat('0',64))
        AND (absence_sha256 IS NULL OR
             pg_catalog.encode(absence_sha256,'hex') IS DISTINCT FROM
                pg_catalog.repeat('0',64))),
    CHECK((object_key IS NULL AND object_version IS NULL
            AND reconcile_provider_request_id IS NULL
            AND reconcile_evidence_sha256 IS NULL
            AND upload_completed_at IS NULL AND reconcile_verified_at IS NULL
            AND delete_provider_request_id IS NULL
            AND delete_response_sha256 IS NULL AND absence_sha256 IS NULL
            AND delete_verified_at IS NULL AND resolution_kind='absent') OR
          (pg_catalog.octet_length(object_key) BETWEEN 1 AND 1024
            AND pg_catalog.octet_length(object_version) BETWEEN 1 AND 512
            AND resolution_kind='deleted'
            AND pg_catalog.octet_length(delete_provider_request_id) BETWEEN 1 AND 512
            AND pg_catalog.octet_length(delete_response_sha256)=32
            AND pg_catalog.octet_length(absence_sha256)=32
            AND delete_verified_at IS NOT NULL)),
    CHECK((reconcile_provider_request_id IS NULL AND
            reconcile_evidence_sha256 IS NULL AND upload_completed_at IS NULL AND
            reconcile_verified_at IS NULL) OR
          (pg_catalog.octet_length(reconcile_provider_request_id) BETWEEN 1 AND 512
            AND pg_catalog.octet_length(reconcile_evidence_sha256)=32
            AND (upload_completed_at IS NULL OR
                 pg_catalog.isfinite(upload_completed_at))
            AND pg_catalog.isfinite(reconcile_verified_at)
            AND (upload_completed_at IS NULL OR
                 upload_completed_at<=reconcile_verified_at))),
    CHECK((delete_provider_request_id IS NULL AND delete_response_sha256 IS NULL
            AND delete_verified_at IS NULL) OR
          (pg_catalog.octet_length(delete_provider_request_id) BETWEEN 1 AND 512
            AND pg_catalog.octet_length(delete_response_sha256)=32
            AND pg_catalog.isfinite(delete_verified_at))),
    CHECK(quota_released_bytes>=0 AND pg_catalog.isfinite(local_state_at)
        AND pg_catalog.isfinite(original_unknown_at)
        AND pg_catalog.isfinite(resolved_at) AND local_state_at<=resolved_at
        AND original_unknown_at<resolved_at
        AND (upload_completed_at IS NULL OR pg_catalog.isfinite(upload_completed_at))
        AND (reconcile_verified_at IS NULL OR reconcile_verified_at<=resolved_at)
        AND (reconcile_verified_at IS NULL OR delete_verified_at IS NULL OR
             reconcile_verified_at<=delete_verified_at)
        AND (delete_verified_at IS NULL OR delete_verified_at<=resolved_at))
);

CREATE TRIGGER analytics_dashboard_export_maintenance_resolution_guard
BEFORE UPDATE OR DELETE
ON analytics_dashboard_export_maintenance_unknown_resolutions
FOR EACH ROW EXECUTE FUNCTION
    analytics_guard_dashboard_export_unknown_resolution_v1();
ALTER TABLE analytics_dashboard_export_maintenance_unknown_resolutions
    ENABLE ALWAYS TRIGGER analytics_dashboard_export_maintenance_resolution_guard;
REVOKE ALL ON TABLE
    analytics_dashboard_export_maintenance_unknown_resolutions FROM PUBLIC;
