-- +goose Up
-- +goose StatementBegin
DO $block$
BEGIN
    IF EXISTS (
        SELECT 1 FROM analytics_dashboard_exports AS item
        WHERE item.state='running' OR item.quota_state='held'
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6105',
            MESSAGE='Analytics dashboard export workers must be drained before evidence adoption';
    END IF;
END;
$block$;

ALTER TABLE analytics_dashboard_exports
    ADD COLUMN platform_attempt_id uuid,
    ADD COLUMN platform_effect_id uuid,
    ADD COLUMN platform_effect_request_sha256 bytea,
    ADD CONSTRAINT analytics_dashboard_exports_platform_evidence_check CHECK (
        (platform_attempt_id IS NULL OR
            pg_catalog.uuid_extract_version(platform_attempt_id)=7)
        AND (platform_effect_id IS NULL OR
            pg_catalog.uuid_extract_version(platform_effect_id)=7)
        AND (platform_effect_request_sha256 IS NULL OR
            pg_catalog.octet_length(platform_effect_request_sha256)=32)
        AND ((platform_effect_id IS NULL AND
              platform_effect_request_sha256 IS NULL) OR
             (platform_effect_id IS NOT NULL AND
              platform_effect_request_sha256 IS NOT NULL))
        AND ((quota_state='held')=(platform_attempt_id IS NOT NULL))
        AND (platform_effect_id IS NULL OR
            (platform_attempt_id IS NOT NULL AND effect_started_at IS NOT NULL))
        AND (effect_started_at IS NULL OR quota_state<>'held' OR
            platform_effect_id IS NOT NULL)
    );

ALTER TABLE analytics_dashboard_export_attempt_receipts
    DROP CONSTRAINT analytics_dashboard_export_attempt_receipts_outcome_check;

ALTER TABLE analytics_dashboard_export_attempt_receipts
    ADD COLUMN platform_attempt_id uuid,
    ADD COLUMN platform_effect_id uuid,
    ADD COLUMN platform_effect_request_sha256 bytea,
    ADD COLUMN platform_effect_receipt_id uuid,
    ADD COLUMN platform_effect_result_sha256 bytea,
    ADD COLUMN uploaded_at timestamptz,
    ADD CONSTRAINT analytics_dashboard_export_receipts_outcome_v2_check CHECK (
        outcome IN ('started','missing','ready','unknown','failed','requeued',
                    'dead','cancelled','deleted')
    ),
    ADD CONSTRAINT analytics_dashboard_export_receipts_platform_evidence_check CHECK (
        (platform_attempt_id IS NULL OR
            pg_catalog.uuid_extract_version(platform_attempt_id)=7)
        AND (platform_effect_id IS NULL OR
            pg_catalog.uuid_extract_version(platform_effect_id)=7)
        AND (platform_effect_receipt_id IS NULL OR
            pg_catalog.uuid_extract_version(platform_effect_receipt_id)=7)
        AND (platform_effect_request_sha256 IS NULL OR
            pg_catalog.octet_length(platform_effect_request_sha256)=32)
        AND (platform_effect_result_sha256 IS NULL OR
            pg_catalog.octet_length(platform_effect_result_sha256)=32)
        AND (uploaded_at IS NULL OR pg_catalog.isfinite(uploaded_at))
        AND ((platform_attempt_id IS NULL AND platform_effect_id IS NULL AND
              platform_effect_request_sha256 IS NULL) OR
             (platform_attempt_id IS NOT NULL AND platform_effect_id IS NOT NULL AND
              platform_effect_request_sha256 IS NOT NULL))
        AND ((platform_effect_receipt_id IS NULL AND
              platform_effect_result_sha256 IS NULL) OR
             (platform_effect_receipt_id IS NOT NULL AND
              platform_effect_id IS NOT NULL AND
              ((outcome='ready' AND
                platform_effect_result_sha256 IS NOT NULL) OR
               (outcome IN ('failed','requeued','dead','cancelled') AND
                platform_effect_result_sha256 IS NULL))))
        AND NOT (platform_effect_id IS NOT NULL AND
            outcome IN ('failed','requeued','dead','cancelled') AND
            platform_effect_receipt_id IS NULL)
        AND NOT (platform_effect_id IS NOT NULL AND outcome='ready' AND
            (platform_effect_receipt_id IS NULL OR
             platform_effect_result_sha256 IS NULL OR uploaded_at IS NULL))
    );

CREATE TABLE analytics_dashboard_export_platform_attempts (
    export_id uuid NOT NULL,
    attempt_generation bigint NOT NULL,
    job_id uuid NOT NULL,
    platform_attempt_id uuid NOT NULL UNIQUE,
    lease_owner uuid NOT NULL,
    claimed_at timestamptz NOT NULL,
    lease_until timestamptz NOT NULL,
    platform_effect_id uuid UNIQUE,
    platform_effect_request_sha256 bytea,
    effect_started_at timestamptz,
    platform_effect_receipt_id uuid UNIQUE,
    platform_effect_result_sha256 bytea,
    uploaded_at timestamptz,
    outcome text NOT NULL,
    failure_class text,
    failure_code text,
    failure_sha256 bytea,
    finished_at timestamptz,
    resolution_receipt_id uuid UNIQUE,
    resolution_kind text,
    resolution_sha256 bytea,
    resolved_at timestamptz,
    PRIMARY KEY(export_id,attempt_generation),
    FOREIGN KEY(export_id) REFERENCES analytics_dashboard_exports(export_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK(attempt_generation BETWEEN 1 AND 3),
    CHECK(pg_catalog.uuid_extract_version(job_id)=7
        AND pg_catalog.uuid_extract_version(platform_attempt_id)=7
        AND lease_owner<>'00000000-0000-0000-0000-000000000000'::uuid),
    CHECK(pg_catalog.isfinite(claimed_at) AND
        pg_catalog.isfinite(lease_until) AND claimed_at<lease_until),
    CHECK(outcome IN ('running','ready','requeued','dead','cancelled','unknown')),
    CHECK((platform_effect_id IS NULL AND
           platform_effect_request_sha256 IS NULL AND effect_started_at IS NULL) OR
          (pg_catalog.uuid_extract_version(platform_effect_id)=7 AND
           pg_catalog.octet_length(platform_effect_request_sha256)=32 AND
           pg_catalog.isfinite(effect_started_at) AND
           effect_started_at>=claimed_at)),
    CHECK((platform_effect_receipt_id IS NULL AND
           platform_effect_result_sha256 IS NULL) OR
          (pg_catalog.uuid_extract_version(platform_effect_receipt_id)=7 AND
           platform_effect_id IS NOT NULL AND
           ((outcome='ready' AND
             pg_catalog.octet_length(platform_effect_result_sha256)=32) OR
             (outcome IN ('requeued','dead','cancelled') AND
              platform_effect_result_sha256 IS NULL)))),
    CHECK((uploaded_at IS NULL AND outcome<>'ready') OR
          (outcome='ready' AND pg_catalog.isfinite(uploaded_at))),
    CHECK((outcome='running' AND finished_at IS NULL AND
           failure_class IS NULL AND failure_code IS NULL AND
           failure_sha256 IS NULL) OR
          (outcome='ready' AND finished_at IS NOT NULL AND
           platform_effect_receipt_id IS NOT NULL AND
           failure_class IS NULL AND failure_code IS NULL AND
           failure_sha256 IS NULL) OR
          (outcome IN ('requeued','dead','cancelled') AND
           finished_at IS NOT NULL AND
           failure_class IN ('transient','permanent','conflict','cancelled') AND
           failure_code ~ '^[A-Z][A-Z0-9_]{1,63}$' AND
           pg_catalog.octet_length(failure_sha256)=32 AND
           (platform_effect_id IS NULL OR
            platform_effect_receipt_id IS NOT NULL)) OR
          (outcome='unknown' AND finished_at IS NOT NULL AND
           platform_effect_id IS NOT NULL AND
           platform_effect_receipt_id IS NULL AND
           failure_class IS NULL AND failure_code='JOB_OUTCOME_UNKNOWN' AND
           failure_sha256 IS NULL)),
    CHECK((resolution_receipt_id IS NULL AND resolution_kind IS NULL AND
           resolution_sha256 IS NULL AND resolved_at IS NULL) OR
          (outcome='unknown' AND
           pg_catalog.uuid_extract_version(resolution_receipt_id)=7 AND
           resolution_kind IN ('deleted','absent') AND
           pg_catalog.octet_length(resolution_sha256)=32 AND
           pg_catalog.isfinite(resolved_at) AND resolved_at>=finished_at))
);
-- +goose StatementEnd
