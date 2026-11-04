-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_v1_actor_assertion_issue_v1(
    in_tenant_id uuid,
    in_domain text
)
RETURNS TABLE (
    iam_group_generation bigint,
    iam_generation bigint,
    target_generation bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        0::bigint,
        0::bigint,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_v1_actor_assertion_issue_v1(uuid, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_v1_actor_assertion_issue_v1(uuid, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_v1_actor_assertion_issue_v1(uuid, text);
