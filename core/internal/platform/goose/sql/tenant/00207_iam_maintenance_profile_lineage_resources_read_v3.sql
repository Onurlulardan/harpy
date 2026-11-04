-- +goose Up
CREATE FUNCTION iam_maintenance_profile_lineage_resources_read_v3(
    in_schema_name name,
    in_tenant_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_payload_revision bigint,
    in_payload_reference text,
    in_digest bytea,
    in_plan_id uuid,
    in_page_number bigint
)
RETURNS TABLE (
    resource_id uuid,
    actual_schema_version_id uuid,
    source_revision bigint,
    document bytea,
    profile_sha256 bytea,
    preconverted boolean,
    planned_target bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::bytea,
        ''::bytea,
        false,
        ''::bytea
    ;
$$;

REVOKE ALL ON FUNCTION iam_maintenance_profile_lineage_resources_read_v3(name, uuid, uuid, uuid, bigint, text, bytea, uuid, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_maintenance_profile_lineage_resources_read_v3(name, uuid, uuid, uuid, bigint, text, bytea, uuid, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_maintenance_profile_lineage_resources_read_v3(name, uuid, uuid, uuid, bigint, text, bytea, uuid, bigint);
