-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_refresh_locator_advance_v1(
    p_selector bytea,
    p_tenant_id uuid,
    p_group_generation bigint,
    p_cutover_generation bigint,
    p_expected_source_revision bigint,
    p_source_revision bigint,
    p_source_sha256 bytea,
    p_source_token_state text,
    p_projected_at timestamptz,
    p_retain_until timestamptz
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_count bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_expected_source_revision IS NULL OR p_source_revision IS NULL OR
       p_source_revision OPERATOR(pg_catalog.<=) p_expected_source_revision THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM refresh locator advance input is invalid';
    END IF;
    PERFORM platform.iam_assert_current_tenant_authority_v1(
        p_tenant_id, p_group_generation, p_cutover_generation
    );

    UPDATE platform.iam_refresh_locators AS locator
    SET source_revision = p_source_revision,
        source_sha256 = p_source_sha256,
        source_token_state = p_source_token_state,
        projected_at = p_projected_at,
        retain_until = p_retain_until
    WHERE locator.selector OPERATOR(pg_catalog.=) p_selector
      AND locator.tenant_id OPERATOR(pg_catalog.=) p_tenant_id
      AND locator.authority_group_generation
          OPERATOR(pg_catalog.=) p_group_generation
      AND locator.cutover_generation
          OPERATOR(pg_catalog.=) p_cutover_generation
      AND locator.source_revision
          OPERATOR(pg_catalog.=) p_expected_source_revision;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    IF v_count IS DISTINCT FROM 1::bigint THEN
        RAISE EXCEPTION USING ERRCODE = 'P5711',
            MESSAGE = 'IAM refresh locator revision is stale or missing';
    END IF;
    RETURN p_source_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_refresh_locator_advance_v1(
    bytea,uuid,bigint,bigint,bigint,bigint,bytea,text,timestamptz,timestamptz
) FROM PUBLIC;
