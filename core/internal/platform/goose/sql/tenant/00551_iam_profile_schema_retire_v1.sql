-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_profile_schema_retire_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_id uuid,
    in_resource text,
    in_version bigint,
    in_definition_json jsonb,
    in_hash bytea,
    in_expected_active_schema_id uuid,
    in_field_id uuid,
    in_expected_field_revision bigint,
    in_manifest_sha256 bytea,
    in_checked_owners integer,
    in_actor_id uuid,
    in_manage_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid,
    result_revision bigint,
    authorization_clock bigint,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        false
    ;
$$;

REVOKE ALL ON FUNCTION iam_profile_schema_retire_v1(name, uuid, bytea, uuid, text, bigint, jsonb, bytea, uuid, uuid, bigint, bytea, integer, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_profile_schema_retire_v1(name, uuid, bytea, uuid, text, bigint, jsonb, bytea, uuid, uuid, bigint, bytea, integer, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_profile_schema_retire_v1(name, uuid, bytea, uuid, text, bigint, jsonb, bytea, uuid, uuid, bigint, bytea, integer, uuid, uuid, bigint, bigint);
