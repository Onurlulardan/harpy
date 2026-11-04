-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_resolve_tenant_target_v1(
    in_tenant_id uuid
)
RETURNS TABLE (
    tenant_id uuid,
    group_generation bigint,
    p_2bytea bytea,
    group_generation_2 bigint,
    iam_generation bigint,
    layout_generation bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        in_tenant_id,
        0::bigint,
        ''::bytea,
        0::bigint,
        0::bigint,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_resolve_tenant_target_v1(uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_resolve_tenant_target_v1(uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_resolve_tenant_target_v1(uuid);
