-- +goose Up
-- BELIRSIZ: no domain generation is stored. Zero rows leave the caller closed.
-- +goose StatementBegin
CREATE FUNCTION platform.iam_snapshot_domain_generations_v1(
    p_tenant_id uuid,
    p_first_domain text,
    p_second_domain text
)
RETURNS TABLE (
    result_domain text,
    result_generation bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT NULL::text, NULL::bigint
    WHERE false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_snapshot_domain_generations_v1(uuid, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_snapshot_domain_generations_v1(uuid, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_snapshot_domain_generations_v1(uuid, text, text);
