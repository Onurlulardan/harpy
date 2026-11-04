-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_profile_schema_read_v2(
    in_schema_name name,
    in_tenant_id uuid,
    in_job_id uuid,
    in_attempt_id uuid,
    in_payload_revision bigint,
    in_payload_reference text,
    in_digest bytea,
    in_plan_id uuid,
    in_id uuid
)
RETURNS TABLE (
    id uuid,
    header_resource text,
    version bigint,
    state text,
    definition_sha bytea,
    field_id uuid,
    key text,
    kind text,
    cardinality text,
    required boolean,
    classification text,
    normalization text,
    state_2 text,
    validation text,
    validation_sha bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        in_id,
        ''::text,
        0::bigint,
        ''::text,
        ''::bytea,
        NULL::uuid,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::boolean,
        NULL::text,
        NULL::text,
        NULL::text,
        NULL::text,
        ''::bytea
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_profile_schema_read_v2(name, uuid, uuid, uuid, bigint, text, bytea, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_maintenance_profile_schema_read_v2(name, uuid, uuid, uuid, bigint, text, bytea, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_maintenance_profile_schema_read_v2(name, uuid, uuid, uuid, bigint, text, bytea, uuid, uuid);
