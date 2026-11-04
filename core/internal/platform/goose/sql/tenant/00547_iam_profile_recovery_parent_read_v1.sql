-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_profile_recovery_parent_read_v1(
    in_schema_name name,
    in_parent_id uuid,
    in_actor_id uuid,
    in_manage_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid,
    recovery_root_plan_id uuid,
    recovery_sequence bigint,
    plan_sha256 bytea,
    resource_kind text,
    source_schema_id uuid,
    target_schema_id uuid,
    converter_version text,
    existing_child_plan_id uuid,
    source_snapshot bytea,
    target_snapshot bytea,
    mappings bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        NULL::uuid,
        0::bigint,
        ''::bytea,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        NULL::uuid,
        ''::bytea,
        ''::bytea,
        ''::bytea
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_recovery_parent_read_v1(name, uuid, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_profile_recovery_parent_read_v1(name, uuid, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_profile_recovery_parent_read_v1(name, uuid, uuid, uuid, bigint, bigint);
