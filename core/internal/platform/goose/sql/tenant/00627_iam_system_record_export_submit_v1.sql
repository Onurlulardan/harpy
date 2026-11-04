-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_submit_v1(
    in_schema_name name,
    in_export_id uuid,
    in_job_id uuid,
    in_receipt_id uuid,
    in_actor_id uuid,
    in_resource text,
    in_token_version bigint,
    in_authorization_version bigint,
    in_authorization_clock bigint,
    in_catalog_version_id uuid,
    in_catalog_sha256 bytea,
    in_query_document bytea,
    in_query_sha256 bytea,
    in_request_sha256 bytea,
    in_idempotency_key_sha256 bytea,
    in_format text,
    in_created_at timestamptz
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
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        in_export_id,
        in_job_id,
        ''::text,
        in_format,
        ''::text,
        0::bigint,
        0::bigint,
        in_created_at,
        NULL::timestamptz,
        NULL::timestamptz,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_submit_v1(name, uuid, uuid, uuid, uuid, text, bigint, bigint, bigint, uuid, bytea, bytea, bytea, bytea, bytea, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_export_submit_v1(name, uuid, uuid, uuid, uuid, text, bigint, bigint, bigint, uuid, bytea, bytea, bytea, bytea, bytea, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_export_submit_v1(name, uuid, uuid, uuid, uuid, text, bigint, bigint, bigint, uuid, bytea, bytea, bytea, bytea, bytea, text, timestamptz);
