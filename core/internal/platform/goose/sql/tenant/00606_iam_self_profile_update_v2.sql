-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_self_profile_update_v2(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_actor_id uuid,
    in_expected_source_schema_id uuid,
    in_schema_id uuid,
    in_canonical_json bytea,
    in_digest bytea,
    in_expected_revision bigint,
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
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_self_profile_update_v2(name, uuid, bytea, uuid, uuid, uuid, bytea, bytea, bigint, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_self_profile_update_v2(name, uuid, bytea, uuid, uuid, uuid, bytea, bytea, bigint, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_self_profile_update_v2(name, uuid, bytea, uuid, uuid, uuid, bytea, bytea, bigint, bigint, bigint);
