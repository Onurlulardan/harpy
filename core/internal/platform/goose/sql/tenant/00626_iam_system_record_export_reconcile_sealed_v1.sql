-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_system_record_export_reconcile_sealed_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_export_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_lease_owner_id uuid,
    in_lease_generation bigint,
    in_resource text,
    in_effect_id uuid,
    in_effect_receipt_id uuid,
    in_request_sha256 bytea,
    in_started_at timestamptz,
    in_object_key text,
    in_object_version text,
    in_key_id text,
    in_provider_request_id text,
    in_effect_id_2 uuid,
    in_effect_receipt_id_2 uuid,
    in_header_sha256 bytea,
    in_ciphertext_sha256 bytea,
    in_request_sha256_2 bytea,
    in_result_sha256 bytea,
    in_ciphertext_bytes bigint,
    in_wrapped_dek bytea,
    in_wrap_nonce bytea,
    in_attempt_generation bigint,
    in_sealed_at timestamptz,
    in_provider_request_id_2 text,
    in_result_sha256_2 bytea,
    in_observed_at timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION iam_system_record_export_reconcile_sealed_v1(name, uuid, uuid, uuid, uuid, uuid, bigint, text, uuid, uuid, bytea, timestamptz, text, text, text, text, uuid, uuid, bytea, bytea, bytea, bytea, bigint, bytea, bytea, bigint, timestamptz, text, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_export_reconcile_sealed_v1(name, uuid, uuid, uuid, uuid, uuid, bigint, text, uuid, uuid, bytea, timestamptz, text, text, text, text, uuid, uuid, bytea, bytea, bytea, bytea, bigint, bytea, bytea, bigint, timestamptz, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_export_reconcile_sealed_v1(name, uuid, uuid, uuid, uuid, uuid, bigint, text, uuid, uuid, bytea, timestamptz, text, text, text, text, uuid, uuid, bytea, bytea, bytea, bytea, bigint, bytea, bytea, bigint, timestamptz, text, bytea, timestamptz);
