-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_resolve_tenant_target_v1(p_tenant_id uuid)
RETURNS TABLE (
    tenant_id uuid,
    group_generation bigint,
    iam_generation bigint,
    layout_generation bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_generation bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_tenant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM tenant target input is invalid';
    END IF;

    SELECT authority.generation INTO v_generation
    FROM platform.cutover_groups AS authority
    WHERE authority.scope_kind OPERATOR(pg_catalog.=) 'tenant'
      AND authority.scope_id OPERATOR(pg_catalog.=) p_tenant_id
      AND authority.group_id OPERATOR(pg_catalog.=) 'iam_system_resources'
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5701',
            MESSAGE = 'IAM tenant target authority is missing';
    END IF;
    PERFORM platform.iam_assert_current_tenant_authority_v1(
        p_tenant_id, v_generation, v_generation
    );

    tenant_id := p_tenant_id;
    group_generation := v_generation;
    iam_generation := v_generation;
    layout_generation := v_generation;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_resolve_tenant_target_v1(uuid) FROM PUBLIC;
