-- +goose Up
CREATE TABLE platform.analytics_dashboard_export_unknown_cleanup_claims (
    job_id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    export_id uuid NOT NULL,
    attempt_id uuid NOT NULL,
    attempt_generation bigint NOT NULL,
    original_lease_owner_id uuid NOT NULL,
    effect_id uuid NOT NULL,
    request_sha256 text NOT NULL,
    claim_generation bigint NOT NULL DEFAULT 0,
    lease_owner_id uuid,
    lease_heartbeat_at timestamptz,
    lease_until timestamptz,
    run_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT analytics_dashboard_export_unknown_claim_tenant_fk
        FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id)
        ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_export_unknown_claim_job_fk
        FOREIGN KEY (job_id) REFERENCES platform.jobs(id) ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_export_unknown_claim_attempt_fk
        FOREIGN KEY (attempt_id) REFERENCES platform.job_attempts(attempt_id)
        ON DELETE RESTRICT,
    CONSTRAINT analytics_dashboard_export_unknown_claim_export_unique
        UNIQUE (tenant_id,export_id),
    CONSTRAINT analytics_dashboard_export_unknown_claim_attempt_unique
        UNIQUE (attempt_id),
    CONSTRAINT analytics_dashboard_export_unknown_claim_effect_unique
        UNIQUE (effect_id),
    CONSTRAINT analytics_dashboard_export_unknown_claim_evidence_check CHECK (
        pg_catalog.uuid_extract_version(export_id)=7 AND
        pg_catalog.uuid_extract_version(job_id)=7 AND
        pg_catalog.uuid_extract_version(attempt_id)=7 AND
        attempt_generation BETWEEN 1 AND 3 AND
        pg_catalog.uuid_extract_version(original_lease_owner_id)=7 AND
        pg_catalog.uuid_extract_version(effect_id)=7 AND
        request_sha256~'^[0-9a-f]{64}$' AND request_sha256<>repeat('0',64)
    ),
    CONSTRAINT analytics_dashboard_export_unknown_claim_lease_check CHECK (
        claim_generation>=0 AND (
            lease_owner_id IS NULL AND lease_heartbeat_at IS NULL AND
            lease_until IS NULL OR
            pg_catalog.uuid_extract_version(lease_owner_id)=7 AND
            claim_generation>0 AND lease_heartbeat_at IS NOT NULL AND
            lease_until IS NOT NULL AND lease_heartbeat_at<lease_until
        )
    ),
    CONSTRAINT analytics_dashboard_export_unknown_claim_time_check CHECK (
        pg_catalog.isfinite(run_at) AND pg_catalog.isfinite(created_at) AND
        pg_catalog.isfinite(updated_at) AND created_at<=updated_at
    )
);

CREATE INDEX analytics_dashboard_export_unknown_claim_due_idx
ON platform.analytics_dashboard_export_unknown_cleanup_claims(
    run_at,lease_until,job_id
);

REVOKE ALL ON TABLE
    platform.analytics_dashboard_export_unknown_cleanup_claims FROM PUBLIC;
