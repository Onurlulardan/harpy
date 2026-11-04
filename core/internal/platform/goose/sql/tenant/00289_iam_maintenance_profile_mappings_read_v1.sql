-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_profile_mappings_read_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_payload_revision bigint,
    in_payload_reference text,
    in_digest bytea,
    in_plan_id uuid
)
RETURNS TABLE (
    source_id uuid,
    target_id uuid,
    kind text,
    semantic_sha bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::bytea
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_profile_mappings_read_v1(name, uuid, uuid, uuid, bigint, text, bytea, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_maintenance_profile_mappings_read_v1(name, uuid, uuid, uuid, bigint, text, bytea, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_maintenance_profile_mappings_read_v1(name, uuid, uuid, uuid, bigint, text, bytea, uuid);
