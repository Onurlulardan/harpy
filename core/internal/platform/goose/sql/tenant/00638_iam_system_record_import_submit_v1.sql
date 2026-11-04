-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_system_record_import_submit_v1(
    in_schema_name name,
    in_import_id uuid,
    in_job_id uuid,
    in_receipt_id uuid,
    in_actor_id uuid,
    in_resource text,
    in_source_id uuid,
    in_token_version bigint,
    in_authorization_version bigint,
    in_authorization_clock bigint,
    in_catalog_version_id uuid,
    in_catalog_sha256 bytea,
    in_identity_manifest_sha256 bytea,
    in_request_sha256 bytea,
    in_idempotency_key_sha256 bytea,
    in_duplicate_action text,
    in_mapping_document bytea,
    in_mapping_sha256 bytea,
    in_field_count bigint,
    in_created_at timestamptz
)
RETURNS TABLE (
    import_id uuid,
    job_id uuid,
    resource_kind text,
    state text,
    processed_rows bigint,
    created_rows bigint,
    updated_rows bigint,
    skipped_rows bigint,
    rejected_rows bigint,
    created_at timestamptz,
    completed_at timestamptz,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_import_id,
        in_job_id,
        ''::text,
        ''::text,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        0::bigint,
        in_created_at,
        NULL::timestamptz,
        false
    ;
$$;

REVOKE ALL ON FUNCTION iam_system_record_import_submit_v1(name, uuid, uuid, uuid, uuid, text, uuid, bigint, bigint, bigint, uuid, bytea, bytea, bytea, bytea, text, bytea, bytea, bigint, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_record_import_submit_v1(name, uuid, uuid, uuid, uuid, text, uuid, bigint, bigint, bigint, uuid, bytea, bytea, bytea, bytea, text, bytea, bytea, bigint, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_record_import_submit_v1(name, uuid, uuid, uuid, uuid, text, uuid, bigint, bigint, bigint, uuid, bytea, bytea, bytea, bytea, text, bytea, bytea, bigint, timestamptz);
