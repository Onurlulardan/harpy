-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_system_record_export_content_v1(
    in_schema_name name,
    in_transfer_id uuid,
    in_actor_id uuid,
    in_resource text,
    in_token_version bigint,
    in_authorization_version bigint,
    in_now timestamptz
)
RETURNS TABLE (
    export_id uuid,
    job_id uuid,
    resource_kind text,
    format text,
    state text,
    processed_rows bigint,
    plaintext_bytes bigint,
    created_at timestamptz,
    ready_at timestamptz,
    expires_at timestamptz,
    object_key text,
    object_version text,
    header_sha256 bytea,
    ciphertext_sha256 bytea,
    ciphertext_bytes bigint,
    key_id text,
    wrapped_dek bytea,
    wrap_nonce bytea,
    attempt_generation bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        0::bigint,
        pg_catalog.clock_timestamp(),
        NULL::timestamptz,
        NULL::timestamptz,
        ''::text,
        ''::text,
        ''::bytea,
        ''::bytea,
        0::bigint,
        ''::text,
        ''::bytea,
        ''::bytea,
        0::bigint
    ;
$$;

REVOKE ALL ON FUNCTION iam_system_record_export_content_v1(name, uuid, uuid, text, bigint, bigint, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_export_content_v1(name, uuid, uuid, text, bigint, bigint, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_export_content_v1(name, uuid, uuid, text, bigint, bigint, timestamptz);
