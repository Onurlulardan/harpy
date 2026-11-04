-- +goose Up
CREATE FUNCTION platform.iam_refresh_locator_create_v1(
    in_selector bytea,
    in_tenant_id uuid,
    in_token_id uuid,
    in_family_id uuid,
    in_authority_group_generation bigint,
    in_iam_generation bigint,
    in_token_revision bigint,
    in_locator_sha256 bytea,
    in_arg text,
    in_issued_at timestamptz,
    in_issued_at_2 timestamptz,
    in_absolute_expires_at timestamptz,
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

REVOKE ALL ON FUNCTION platform.iam_refresh_locator_create_v1(bytea, uuid, uuid, uuid, bigint, bigint, bigint, bytea, text, timestamptz, timestamptz, timestamptz, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_refresh_locator_create_v1(bytea, uuid, uuid, uuid, bigint, bigint, bigint, bytea, text, timestamptz, timestamptz, timestamptz, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_refresh_locator_create_v1(bytea, uuid, uuid, uuid, bigint, bigint, bigint, bytea, text, timestamptz, timestamptz, timestamptz, timestamptz);
