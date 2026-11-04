-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_system_record_import_worker_claim_v2(
    in_schema_name name,
    in_tenant_id uuid,
    in_import_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_lease_owner_id uuid,
    in_lease_generation bigint
)
RETURNS TABLE (
    actor_id uuid,
    resource_kind text,
    authorization_version bigint,
    duplicate_action text,
    mapping_document bytea,
    next_row_index bigint,
    processed_rows bigint,
    source_id uuid,
    source_format text,
    object_key text,
    object_version text,
    header_sha256 bytea,
    ciphertext_sha256 bytea,
    ciphertext_bytes bigint,
    plaintext_sha256 bytea,
    plaintext_bytes bigint,
    key_id text,
    wrapped_dek bytea,
    wrap_nonce bytea,
    provider_request_id text,
    uploaded_at timestamptz,
    finished boolean,
    catalog_version_id uuid,
    catalog_sha256 bytea,
    identity_manifest_sha256 bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        0::bigint,
        ''::text,
        ''::bytea,
        0::bigint,
        0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        ''::text,
        ''::bytea,
        ''::bytea,
        0::bigint,
        ''::bytea,
        0::bigint,
        ''::text,
        ''::bytea,
        ''::bytea,
        ''::text,
        pg_catalog.clock_timestamp(),
        false,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::bytea,
        ''::bytea
    ;
$$;

REVOKE ALL ON FUNCTION iam_system_record_import_worker_claim_v2(name, uuid, uuid, uuid, uuid, uuid, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_import_worker_claim_v2(name, uuid, uuid, uuid, uuid, uuid, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_import_worker_claim_v2(name, uuid, uuid, uuid, uuid, uuid, bigint);
