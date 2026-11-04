-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_reconcile_claim_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_export_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_lease_owner_id uuid,
    in_lease_generation bigint,
    in_resource text
)
RETURNS TABLE (
    claimed boolean,
    actor_id uuid,
    original_job_id uuid,
    token bigint,
    "authorization" bigint,
    catalog_version_id uuid,
    catalog_sha bytea,
    query_document bytea,
    query_sha bytea,
    format text,
    original_generation bigint,
    original_effect_id uuid,
    original_receipt_id uuid,
    original_lease_owner_id uuid,
    object_key text,
    original_request_sha bytea,
    original_started_at timestamptz,
    key_id text,
    header_sha bytea,
    ciphertext_sha bytea,
    ciphertext_bytes bigint,
    wrapped_dek bytea,
    wrap_nonce bytea,
    observation_effect_id uuid,
    observation_receipt_id uuid,
    observation_request_sha bytea,
    observation_started_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        false,
        NULL::uuid,
        NULL::uuid,
        NULL::bigint,
        NULL::bigint,
        NULL::uuid,
        ''::bytea,
        ''::bytea,
        ''::bytea,
        NULL::text,
        NULL::bigint,
        NULL::uuid,
        NULL::uuid,
        NULL::uuid,
        NULL::text,
        ''::bytea,
        NULL::timestamptz,
        NULL::text,
        ''::bytea,
        ''::bytea,
        NULL::bigint,
        ''::bytea,
        ''::bytea,
        NULL::uuid,
        NULL::uuid,
        ''::bytea,
        NULL::timestamptz
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_reconcile_claim_v1(name, uuid, uuid, uuid, uuid, uuid, bigint, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_export_reconcile_claim_v1(name, uuid, uuid, uuid, uuid, uuid, bigint, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_export_reconcile_claim_v1(name, uuid, uuid, uuid, uuid, uuid, bigint, text);
