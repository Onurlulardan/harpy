-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_worker_complete_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_actor_id uuid,
    in_export_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_lease_owner_id uuid,
    in_attempt_generation bigint,
    in_resource text,
    in_token_version bigint,
    in_authorization_version bigint,
    in_rows bigint,
    in_plaintext_bytes bigint,
    in_checked_at timestamptz,
    in_object_key text,
    in_object_version text,
    in_key_id text,
    in_provider_request_id text,
    in_effect_id uuid,
    in_effect_receipt_id uuid,
    in_header_sha256 bytea,
    in_ciphertext_sha256 bytea,
    in_request_sha256 bytea,
    in_result_sha256 bytea,
    in_ciphertext_bytes bigint,
    in_wrapped_dek bytea,
    in_wrap_nonce bytea,
    in_sealed_at timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_worker_complete_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, bigint, bigint, timestamptz, text, text, text, text, uuid, uuid, bytea, bytea, bytea, bytea, bigint, bytea, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_export_worker_complete_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, bigint, bigint, timestamptz, text, text, text, text, uuid, uuid, bytea, bytea, bytea, bytea, bigint, bytea, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_export_worker_complete_v1(name, uuid, uuid, uuid, uuid, uuid, uuid, bigint, text, bigint, bigint, bigint, bigint, timestamptz, text, text, text, text, uuid, uuid, bytea, bytea, bytea, bytea, bigint, bytea, bytea, timestamptz);
