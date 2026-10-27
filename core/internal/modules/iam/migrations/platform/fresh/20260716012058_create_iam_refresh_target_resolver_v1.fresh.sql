-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_resolve_refresh_target_v1(p_selector bytea)
RETURNS TABLE (
    tenant_id uuid,
    source_revision bigint,
    source_sha256 bytea,
    group_generation bigint,
    iam_generation bigint,
    layout_generation bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_probe_tenant_id uuid;
    v_locked_group_generation bigint;
    v_locked_iam_generation bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_selector IS NULL OR
       pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM refresh target input is invalid';
    END IF;

    SELECT locator.tenant_id INTO v_probe_tenant_id
    FROM platform.iam_refresh_locators AS locator
    WHERE locator.selector OPERATOR(pg_catalog.=) p_selector
      AND locator.retain_until OPERATOR(pg_catalog.>) pg_catalog.clock_timestamp();
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5732',
            MESSAGE = 'IAM refresh target is not retained';
    END IF;

    SELECT target.group_generation, target.iam_generation,
           target.layout_generation
    INTO group_generation, iam_generation, layout_generation
    FROM platform.iam_resolve_tenant_target_v1(v_probe_tenant_id) AS target;

    SELECT locator.tenant_id, locator.source_revision, locator.source_sha256,
           locator.authority_group_generation, locator.cutover_generation
    INTO tenant_id, source_revision, source_sha256,
         v_locked_group_generation, v_locked_iam_generation
    FROM platform.iam_refresh_locators AS locator
    WHERE locator.selector OPERATOR(pg_catalog.=) p_selector
      AND locator.retain_until OPERATOR(pg_catalog.>) pg_catalog.clock_timestamp()
    FOR SHARE;
    IF NOT FOUND OR tenant_id IS DISTINCT FROM v_probe_tenant_id OR
       v_locked_group_generation IS DISTINCT FROM group_generation OR
       v_locked_iam_generation IS DISTINCT FROM iam_generation THEN
        RAISE EXCEPTION USING ERRCODE = 'P5732',
            MESSAGE = 'IAM refresh target authority is stale or changed';
    END IF;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_resolve_refresh_target_v1(bytea) FROM PUBLIC;
