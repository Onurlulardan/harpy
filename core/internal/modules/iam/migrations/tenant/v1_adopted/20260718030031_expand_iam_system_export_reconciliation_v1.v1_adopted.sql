-- +goose Up
ALTER TABLE iam_system_record_exports
    ADD COLUMN reconciliation_job_id uuid,
    ADD COLUMN reconciliation_state text,
    ADD COLUMN reconciliation_attempt_generation bigint,
    ADD COLUMN reconciliation_effect_id uuid,
    ADD COLUMN reconciliation_receipt_id uuid,
    ADD COLUMN reconciliation_request_sha256 bytea,
    ADD COLUMN reconciliation_started_at timestamptz,
    ADD COLUMN reconciliation_observed_at timestamptz,
    ADD COLUMN reconciliation_next_at timestamptz,
    ADD COLUMN reconciliation_code text;

ALTER TABLE iam_system_record_exports
    ADD CONSTRAINT iam_system_record_exports_reconciliation_check CHECK (
        (reconciliation_job_id IS NULL AND reconciliation_state IS NULL AND
         reconciliation_attempt_generation IS NULL AND
         reconciliation_effect_id IS NULL AND reconciliation_receipt_id IS NULL AND
         reconciliation_request_sha256 IS NULL AND
         reconciliation_started_at IS NULL AND reconciliation_observed_at IS NULL AND
         reconciliation_next_at IS NULL AND reconciliation_code IS NULL) OR
        (pg_catalog.uuid_extract_version(reconciliation_job_id) = 7 AND
         ((state='outcome_unknown' AND reconciliation_state='queued' AND
           reconciliation_attempt_generation IS NULL AND
           reconciliation_effect_id IS NULL AND reconciliation_receipt_id IS NULL AND
           reconciliation_request_sha256 IS NULL AND
           reconciliation_started_at IS NULL AND reconciliation_observed_at IS NULL AND
           reconciliation_next_at IS NULL AND reconciliation_code IS NULL) OR
          (state='outcome_unknown' AND reconciliation_state='running' AND
           reconciliation_attempt_generation BETWEEN 1 AND 8 AND
           pg_catalog.uuid_extract_version(reconciliation_effect_id) = 7 AND
           pg_catalog.uuid_extract_version(reconciliation_receipt_id) = 7 AND
           reconciliation_effect_id<>reconciliation_receipt_id AND
           pg_catalog.octet_length(reconciliation_request_sha256) = 32 AND
           pg_catalog.isfinite(reconciliation_started_at) AND
           reconciliation_observed_at IS NULL AND reconciliation_next_at IS NULL AND
           reconciliation_code IS NULL) OR
          (state='outcome_unknown' AND reconciliation_state='retry_wait' AND
           reconciliation_attempt_generation BETWEEN 1 AND 7 AND
           pg_catalog.uuid_extract_version(reconciliation_effect_id) = 7 AND
           pg_catalog.uuid_extract_version(reconciliation_receipt_id) = 7 AND
           reconciliation_effect_id<>reconciliation_receipt_id AND
           pg_catalog.octet_length(reconciliation_request_sha256) = 32 AND
           pg_catalog.isfinite(reconciliation_started_at) AND
           pg_catalog.isfinite(reconciliation_observed_at) AND
           reconciliation_next_at>=reconciliation_observed_at AND
           reconciliation_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN') OR
          (state='outcome_unknown' AND reconciliation_state='parked' AND
           reconciliation_attempt_generation BETWEEN 1 AND 8 AND
           pg_catalog.uuid_extract_version(reconciliation_effect_id) = 7 AND
           pg_catalog.uuid_extract_version(reconciliation_receipt_id) = 7 AND
           reconciliation_effect_id<>reconciliation_receipt_id AND
           pg_catalog.octet_length(reconciliation_request_sha256) = 32 AND
           pg_catalog.isfinite(reconciliation_started_at) AND
           pg_catalog.isfinite(reconciliation_observed_at) AND
           reconciliation_next_at IS NULL AND
           reconciliation_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_PARKED') OR
          (state='ready' AND reconciliation_state='resolved_sealed' AND
           reconciliation_attempt_generation BETWEEN 1 AND 8 AND
           pg_catalog.uuid_extract_version(reconciliation_effect_id) = 7 AND
           pg_catalog.uuid_extract_version(reconciliation_receipt_id) = 7 AND
           reconciliation_effect_id<>reconciliation_receipt_id AND
           pg_catalog.octet_length(reconciliation_request_sha256) = 32 AND
           pg_catalog.isfinite(reconciliation_started_at) AND
           pg_catalog.isfinite(reconciliation_observed_at) AND
           reconciliation_next_at IS NULL AND
           reconciliation_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILED') OR
          (state='failed' AND reconciliation_state='resolved_missing' AND
           reconciliation_attempt_generation BETWEEN 1 AND 8 AND
           pg_catalog.uuid_extract_version(reconciliation_effect_id) = 7 AND
           pg_catalog.uuid_extract_version(reconciliation_receipt_id) = 7 AND
           reconciliation_effect_id<>reconciliation_receipt_id AND
           pg_catalog.octet_length(reconciliation_request_sha256) = 32 AND
           pg_catalog.isfinite(reconciliation_started_at) AND
           pg_catalog.isfinite(reconciliation_observed_at) AND
           reconciliation_next_at IS NULL AND
           reconciliation_code='IAM_SYSTEM_EXPORT_ARTIFACT_MISSING')))
    );

CREATE UNIQUE INDEX iam_system_record_exports_reconciliation_job_uidx
ON iam_system_record_exports (reconciliation_job_id)
WHERE reconciliation_job_id IS NOT NULL;
