-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_portal_setup_read_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_binding_id uuid,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    binding_id uuid,
    user_id uuid,
    setup_state text,
    approval_state text,
    revision bigint,
    setup_completed_at timestamptz,
    submitted_at timestamptz,
    decision_at timestamptz,
    updated_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        in_binding_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        0::bigint,
        NULL::timestamptz,
        NULL::timestamptz,
        NULL::timestamptz,
        pg_catalog.clock_timestamp()
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_portal_setup_read_v1(name, uuid, uuid, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_portal_setup_read_v1(name, uuid, uuid, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_portal_setup_read_v1(name, uuid, uuid, uuid, uuid, bigint, bigint);
