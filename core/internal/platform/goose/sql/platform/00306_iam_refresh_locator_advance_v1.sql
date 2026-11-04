-- +goose Up
CREATE FUNCTION platform.iam_refresh_locator_advance_v1(
    in_selector bytea,
    in_tenant_id uuid,
    in_authority_group_generation bigint,
    in_iam_generation bigint,
    in_expected bigint,
    in_revision bigint,
    in_source_sha256 bytea,
    in_state text,
    in_projected_at timestamptz,
    in_retain_until timestamptz
)
RETURNS void
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT NULL::void;
$$;

REVOKE ALL ON FUNCTION platform.iam_refresh_locator_advance_v1(bytea, uuid, bigint, bigint, bigint, bigint, bytea, text, timestamptz, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_refresh_locator_advance_v1(bytea, uuid, bigint, bigint, bigint, bigint, bytea, text, timestamptz, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_refresh_locator_advance_v1(bytea, uuid, bigint, bigint, bigint, bigint, bytea, text, timestamptz, timestamptz);
