-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_group_create_v1(
    in_schema_name name,
    in_group_id uuid,
    in_requester_id uuid,
    in_mode text,
    in_manage uuid,
    in_review uuid,
    in_expected_authorization_version bigint,
    in_step_up_purpose_sha256 bytea,
    in_expires_at timestamptz,
    in_payload jsonb
)
RETURNS TABLE (
    group_id uuid,
    group_state text,
    group_revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        in_group_id,
        ''::text,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_group_create_v1(name, uuid, uuid, text, uuid, uuid, bigint, bytea, timestamptz, jsonb) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_permission_request_group_create_v1(name, uuid, uuid, text, uuid, uuid, bigint, bytea, timestamptz, jsonb) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_permission_request_group_create_v1(name, uuid, uuid, text, uuid, uuid, bigint, bytea, timestamptz, jsonb);
