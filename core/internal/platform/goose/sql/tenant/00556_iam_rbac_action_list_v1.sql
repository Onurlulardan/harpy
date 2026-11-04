-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_action_list_v1(
    in_schema_name name,
    in_state text,
    in_limit integer,
    in_actor_id uuid,
    in_required_permission_id uuid
)
RETURNS TABLE (
    id uuid,
    key text,
    display_name text,
    description text,
    protected boolean,
    state text,
    revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        NULL::text,
        false,
        in_state,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_action_list_v1(name, text, integer, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_rbac_action_list_v1(name, text, integer, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_rbac_action_list_v1(name, text, integer, uuid, uuid);
