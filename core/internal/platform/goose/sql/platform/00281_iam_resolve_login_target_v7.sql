-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_resolve_login_target_v7(
    in_tenant_id uuid,
    in_kind text,
    in_key_id text,
    in_key bytea
)
RETURNS TABLE (
    tenant_id uuid,
    source_user_id uuid,
    source_identifier_id uuid,
    source_identifier_kind text,
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
        in_tenant_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        0::bigint,
        ''::bytea,
        0::bigint,
        0::bigint,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_resolve_login_target_v7(uuid, text, text, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_resolve_login_target_v7(uuid, text, text, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_resolve_login_target_v7(uuid, text, text, bytea);
