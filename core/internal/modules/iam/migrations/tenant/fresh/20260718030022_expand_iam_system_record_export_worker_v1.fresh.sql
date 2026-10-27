-- +goose Up
ALTER TABLE iam_system_record_exports
    ADD COLUMN failure_code text,
    ADD COLUMN effect_id uuid,
    ADD COLUMN effect_lease_owner_id uuid,
    ADD COLUMN effect_object_key text,
    ADD COLUMN effect_request_sha256 bytea,
    ADD COLUMN effect_started_at timestamptz,
    ADD COLUMN effect_receipt_id uuid,
    ADD COLUMN provider_request_id text,
    ADD COLUMN effect_result_sha256 bytea,
    ADD COLUMN effect_completed_at timestamptz,
    ADD COLUMN outcome_receipt_id uuid;

ALTER TABLE iam_system_record_exports
    DROP CONSTRAINT iam_system_record_exports_artifact_check;

ALTER TABLE iam_system_record_exports
    ADD CONSTRAINT iam_system_record_exports_artifact_check CHECK (
        (state = 'waiting' AND
         object_key IS NULL AND object_version IS NULL AND header_sha256 IS NULL AND
         ciphertext_sha256 IS NULL AND ciphertext_bytes IS NULL AND key_id IS NULL AND
         wrapped_dek IS NULL AND wrap_nonce IS NULL AND attempt_generation IS NULL AND
         ready_at IS NULL AND expires_at IS NULL) OR
        (state = 'running' AND
         object_key IS NULL AND object_version IS NULL AND
         attempt_generation BETWEEN 1 AND 9007199254740991 AND
         ready_at IS NULL AND expires_at IS NULL AND
         ((effect_id IS NULL AND header_sha256 IS NULL AND
           ciphertext_sha256 IS NULL AND ciphertext_bytes IS NULL AND key_id IS NULL AND
           wrapped_dek IS NULL AND wrap_nonce IS NULL) OR
          (effect_id IS NOT NULL AND pg_catalog.octet_length(header_sha256) = 32 AND
           pg_catalog.octet_length(ciphertext_sha256) = 32 AND
           ciphertext_bytes BETWEEN 1 AND 2148532224 AND
           pg_catalog.octet_length(key_id) BETWEEN 1 AND 128 AND
           pg_catalog.octet_length(wrapped_dek) BETWEEN 32 AND 1024 AND
           pg_catalog.octet_length(wrap_nonce) = 12))) OR
        (state = 'failed' AND
         object_key IS NULL AND object_version IS NULL AND
         (attempt_generation IS NULL OR
          attempt_generation BETWEEN 1 AND 9007199254740991) AND
         ready_at IS NULL AND expires_at IS NULL AND
         ((header_sha256 IS NULL AND ciphertext_sha256 IS NULL AND
           ciphertext_bytes IS NULL AND key_id IS NULL AND wrapped_dek IS NULL AND
           wrap_nonce IS NULL) OR
          (effect_id IS NOT NULL AND pg_catalog.octet_length(header_sha256) = 32 AND
           pg_catalog.octet_length(ciphertext_sha256) = 32 AND
           ciphertext_bytes BETWEEN 1 AND 2148532224 AND
           pg_catalog.octet_length(key_id) BETWEEN 1 AND 128 AND
           pg_catalog.octet_length(wrapped_dek) BETWEEN 32 AND 1024 AND
           pg_catalog.octet_length(wrap_nonce) = 12))) OR
        (state = 'cancelled' AND object_key IS NULL AND object_version IS NULL AND
         (attempt_generation IS NULL OR
          attempt_generation BETWEEN 1 AND 9007199254740991) AND
         ready_at IS NULL AND expires_at IS NULL AND
         ((header_sha256 IS NULL AND ciphertext_sha256 IS NULL AND
           ciphertext_bytes IS NULL AND key_id IS NULL AND wrapped_dek IS NULL AND
           wrap_nonce IS NULL) OR
          (effect_id IS NOT NULL AND pg_catalog.octet_length(header_sha256) = 32 AND
           pg_catalog.octet_length(ciphertext_sha256) = 32 AND
           ciphertext_bytes BETWEEN 1 AND 2148532224 AND
           pg_catalog.octet_length(key_id) BETWEEN 1 AND 128 AND
           pg_catalog.octet_length(wrapped_dek) BETWEEN 32 AND 1024 AND
           pg_catalog.octet_length(wrap_nonce) = 12))) OR
        (state = 'outcome_unknown' AND
         ((object_key IS NULL AND object_version IS NULL) OR
          (pg_catalog.octet_length(object_key) BETWEEN 1 AND 512 AND
           pg_catalog.octet_length(object_version) BETWEEN 1 AND 512)) AND
         pg_catalog.octet_length(header_sha256) = 32 AND
         pg_catalog.octet_length(ciphertext_sha256) = 32 AND
         ciphertext_bytes BETWEEN 1 AND 2148532224 AND
         pg_catalog.octet_length(key_id) BETWEEN 1 AND 128 AND
         pg_catalog.octet_length(wrapped_dek) BETWEEN 32 AND 1024 AND
         pg_catalog.octet_length(wrap_nonce) = 12 AND
         attempt_generation BETWEEN 1 AND 9007199254740991 AND
         ready_at IS NULL AND expires_at IS NULL) OR
        (state IN ('ready','expired') AND
         pg_catalog.octet_length(object_key) BETWEEN 1 AND 512 AND
         object_key=effect_object_key AND
         pg_catalog.octet_length(object_version) BETWEEN 1 AND 512 AND
         pg_catalog.octet_length(header_sha256) = 32 AND
         pg_catalog.octet_length(ciphertext_sha256) = 32 AND
         ciphertext_bytes BETWEEN 1 AND 2148532224 AND
         pg_catalog.octet_length(key_id) BETWEEN 1 AND 128 AND
         pg_catalog.octet_length(wrapped_dek) BETWEEN 32 AND 1024 AND
         pg_catalog.octet_length(wrap_nonce) = 12 AND
         attempt_generation BETWEEN 1 AND 9007199254740991 AND
         ready_at IS NOT NULL AND expires_at = ready_at + interval '24 hours')
    );

ALTER TABLE iam_system_record_exports
    ADD CONSTRAINT iam_system_record_exports_failure_check CHECK (
        (state = 'failed' AND failure_code ~ '^[A-Z][A-Z0-9_]{1,63}$') OR
        (state = 'cancelled' AND
         (failure_code IS NULL OR failure_code = 'IAM_SYSTEM_EXPORT_CANCELLED')) OR
        (state = 'outcome_unknown' AND failure_code = 'JOB_OUTCOME_UNKNOWN') OR
        (state NOT IN ('failed','cancelled','outcome_unknown') AND failure_code IS NULL)
    );

ALTER TABLE iam_system_record_exports
    ADD CONSTRAINT iam_system_record_exports_effect_check CHECK (
        (state = 'waiting' AND effect_id IS NULL AND effect_lease_owner_id IS NULL AND
         effect_object_key IS NULL AND
         effect_request_sha256 IS NULL AND effect_started_at IS NULL AND
         effect_receipt_id IS NULL AND provider_request_id IS NULL AND
         effect_result_sha256 IS NULL AND effect_completed_at IS NULL AND
         outcome_receipt_id IS NULL) OR
        (state = 'running' AND
         provider_request_id IS NULL AND
         effect_result_sha256 IS NULL AND effect_completed_at IS NULL AND
         outcome_receipt_id IS NULL AND
          ((effect_id IS NULL AND effect_object_key IS NULL AND
           effect_lease_owner_id IS NULL AND
           effect_request_sha256 IS NULL AND effect_started_at IS NULL AND
           effect_receipt_id IS NULL) OR
          (pg_catalog.uuid_extract_version(effect_id) = 7 AND
           pg_catalog.uuid_extract_version(effect_lease_owner_id) = 7 AND
           pg_catalog.octet_length(effect_object_key) BETWEEN 1 AND 512 AND
           pg_catalog.octet_length(effect_request_sha256) = 32 AND
           pg_catalog.isfinite(effect_started_at) AND
           pg_catalog.uuid_extract_version(effect_receipt_id) = 7))) OR
        (state IN ('failed','cancelled') AND
         provider_request_id IS NULL AND effect_result_sha256 IS NULL AND
         effect_completed_at IS NULL AND
         outcome_receipt_id IS NULL AND
         ((effect_id IS NULL AND effect_object_key IS NULL AND
           effect_lease_owner_id IS NULL AND
           effect_request_sha256 IS NULL AND effect_started_at IS NULL AND
           effect_receipt_id IS NULL) OR
          (pg_catalog.uuid_extract_version(effect_id) = 7 AND
           pg_catalog.uuid_extract_version(effect_lease_owner_id) = 7 AND
           pg_catalog.octet_length(effect_object_key) BETWEEN 1 AND 512 AND
           pg_catalog.octet_length(effect_request_sha256) = 32 AND
           pg_catalog.isfinite(effect_started_at) AND
           (effect_receipt_id IS NULL OR
            (state='cancelled' AND
             pg_catalog.uuid_extract_version(effect_receipt_id) = 7))))) OR
        (state IN ('ready','expired') AND
         pg_catalog.uuid_extract_version(effect_id) = 7 AND
         pg_catalog.uuid_extract_version(effect_lease_owner_id) = 7 AND
         pg_catalog.octet_length(effect_object_key) BETWEEN 1 AND 512 AND
         pg_catalog.octet_length(effect_request_sha256) = 32 AND
         pg_catalog.isfinite(effect_started_at) AND
         pg_catalog.uuid_extract_version(effect_receipt_id) = 7 AND
         pg_catalog.octet_length(provider_request_id) BETWEEN 1 AND 512 AND
         pg_catalog.octet_length(effect_result_sha256) = 32 AND
         ready_at BETWEEN effect_completed_at AND
             effect_completed_at+interval '1 second' AND
         outcome_receipt_id IS NULL) OR
        (state = 'outcome_unknown' AND
         pg_catalog.uuid_extract_version(effect_id) = 7 AND
         pg_catalog.uuid_extract_version(effect_lease_owner_id) = 7 AND
         pg_catalog.octet_length(effect_object_key) BETWEEN 1 AND 512 AND
         pg_catalog.octet_length(effect_request_sha256) = 32 AND
         pg_catalog.isfinite(effect_started_at) AND effect_receipt_id IS NULL AND
         pg_catalog.uuid_extract_version(outcome_receipt_id) = 7 AND
         outcome_receipt_id<>effect_id AND
         ((provider_request_id IS NULL AND effect_result_sha256 IS NULL AND
           effect_completed_at IS NULL AND object_key IS NULL) OR
          (pg_catalog.octet_length(provider_request_id) BETWEEN 1 AND 512 AND
           pg_catalog.octet_length(effect_result_sha256) = 32 AND
           pg_catalog.isfinite(effect_completed_at) AND
           object_key=effect_object_key)))
    );

ALTER TABLE iam_system_record_exports
    ADD CONSTRAINT iam_system_record_exports_effect_time_check CHECK (
        effect_started_at IS NULL OR
         (started_at IS NOT NULL AND effect_started_at >= started_at AND
         (effect_completed_at IS NULL OR
          effect_completed_at >= effect_started_at-interval '1 second') AND
         (completed_at IS NULL OR completed_at >= effect_started_at) AND
         (effect_completed_at IS NULL OR completed_at >= effect_completed_at))
    );
