-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_resolve_refresh_target_v2(
    in_selector bytea
)
RETURNS TABLE (
    tenant_id uuid,
    source_token_id uuid,
    source_family_id uuid,
    source_selector bytea,
    source_revision bigint,
    source_sha256 bytea,
    group_generation bigint,
    iam_generation bigint,
    layout_generation bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::bytea,
        0::bigint,
        ''::bytea,
        0::bigint,
        0::bigint,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_resolve_refresh_target_v2(bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_resolve_refresh_target_v2(bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_resolve_refresh_target_v2(bytea);
