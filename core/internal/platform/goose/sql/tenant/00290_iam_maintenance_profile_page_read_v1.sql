-- +goose Up
CREATE FUNCTION iam_maintenance_profile_page_read_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_payload_revision bigint,
    in_payload_reference text,
    in_digest bytea
)
RETURNS TABLE (
    plan_id uuid,
    resource text,
    source_schema_id uuid,
    target_schema_id uuid,
    converter text,
    state text,
    page_number bigint,
    after uuid,
    through_id uuid,
    expected_count bigint,
    source_sha bytea,
    target_sha bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        0::bigint,
        NULL::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::bytea,
        ''::bytea
    ;
$$;

REVOKE ALL ON FUNCTION iam_maintenance_profile_page_read_v1(name, uuid, uuid, uuid, bigint, text, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_maintenance_profile_page_read_v1(name, uuid, uuid, uuid, bigint, text, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_maintenance_profile_page_read_v1(name, uuid, uuid, uuid, bigint, text, bytea);
