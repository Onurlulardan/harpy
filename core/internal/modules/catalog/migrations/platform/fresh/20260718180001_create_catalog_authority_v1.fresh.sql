-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.catalog_assert_tenant_authority_v1(
    expected_tenant uuid,
    expected_generation bigint
) RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $$
DECLARE
    matched boolean;
BEGIN
    SELECT true INTO matched
    FROM platform.domain_cutovers
    WHERE scope_kind = 'tenant'
      AND scope_id = expected_tenant
      AND domain = 'catalog'
      AND state = 'v2'
      AND writer = 'v2'
      AND generation = expected_generation
    FOR SHARE;
    IF matched IS DISTINCT FROM true OR session_user <> 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'catalog authority is not current';
    END IF;
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.catalog_assert_tenant_authority_v1(uuid,bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.catalog_assert_tenant_authority_v1(uuid,bigint)
    TO forgecrud_runtime;
