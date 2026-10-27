-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.iam_assert_current_tenant_authority_v1(
    p_tenant_id uuid,
    p_group_generation bigint,
    p_cutover_generation bigint
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_group_state text;
    v_group_writer text;
    v_group_generation bigint;
    v_domain_state text;
    v_domain_writer text;
    v_domain_generation bigint;
    v_layout_state text;
    v_layout_writer text;
    v_layout_generation bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_tenant_id IS NULL OR p_group_generation IS NULL OR
       p_cutover_generation IS NULL OR
       p_group_generation OPERATOR(pg_catalog.<) 1 OR
       p_cutover_generation OPERATOR(pg_catalog.<) 1 OR
       p_group_generation IS DISTINCT FROM p_cutover_generation THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM tenant authority input is invalid';
    END IF;

    SELECT authority.state, authority.writer, authority.generation
    INTO v_group_state, v_group_writer, v_group_generation
    FROM platform.cutover_groups AS authority
    WHERE authority.scope_kind OPERATOR(pg_catalog.=) 'tenant'
      AND authority.scope_id OPERATOR(pg_catalog.=) p_tenant_id
      AND authority.group_id OPERATOR(pg_catalog.=) 'iam_system_resources';
    IF NOT FOUND OR v_group_state IS DISTINCT FROM 'v2' OR
       v_group_writer IS DISTINCT FROM 'v2' OR
       v_group_generation IS DISTINCT FROM p_group_generation THEN
        RAISE EXCEPTION USING ERRCODE = 'P5701',
            MESSAGE = 'IAM cutover group authority is not current V2';
    END IF;

    SELECT authority.state, authority.writer, authority.generation
    INTO v_domain_state, v_domain_writer, v_domain_generation
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind OPERATOR(pg_catalog.=) 'tenant'
      AND authority.scope_id OPERATOR(pg_catalog.=) p_tenant_id
      AND authority.domain OPERATOR(pg_catalog.=) 'iam';
    IF NOT FOUND OR v_domain_state IS DISTINCT FROM 'v2' OR
       v_domain_writer IS DISTINCT FROM 'v2' OR
       v_domain_generation IS DISTINCT FROM p_cutover_generation OR
       v_domain_generation IS DISTINCT FROM v_group_generation THEN
        RAISE EXCEPTION USING ERRCODE = 'P5702',
            MESSAGE = 'IAM domain authority is not current V2';
    END IF;

    SELECT authority.state, authority.writer, authority.generation
    INTO v_layout_state, v_layout_writer, v_layout_generation
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind OPERATOR(pg_catalog.=) 'tenant'
      AND authority.scope_id OPERATOR(pg_catalog.=) p_tenant_id
      AND authority.domain OPERATOR(pg_catalog.=) 'system_resource_layout';
    IF NOT FOUND OR v_layout_state IS DISTINCT FROM 'v2' OR
       v_layout_writer IS DISTINCT FROM 'v2' OR
       v_layout_generation IS DISTINCT FROM v_group_generation THEN
        RAISE EXCEPTION USING ERRCODE = 'P5703',
            MESSAGE = 'IAM layout domain authority is not current V2';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM platform.cutover_group_members AS member
        WHERE member.scope_kind OPERATOR(pg_catalog.=) 'tenant'
          AND member.scope_id OPERATOR(pg_catalog.=) p_tenant_id
          AND member.group_id OPERATOR(pg_catalog.=) 'iam_system_resources'
          AND member.ordinal OPERATOR(pg_catalog.=) 1
          AND member.domain OPERATOR(pg_catalog.=) 'iam'
    ) OR NOT EXISTS (
        SELECT 1 FROM platform.cutover_group_members AS member
        WHERE member.scope_kind OPERATOR(pg_catalog.=) 'tenant'
          AND member.scope_id OPERATOR(pg_catalog.=) p_tenant_id
          AND member.group_id OPERATOR(pg_catalog.=) 'iam_system_resources'
          AND member.ordinal OPERATOR(pg_catalog.=) 2
          AND member.domain OPERATOR(pg_catalog.=) 'system_resource_layout'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5704',
            MESSAGE = 'IAM system-resource group membership is invalid';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_assert_current_tenant_authority_v1(
    uuid,bigint,bigint
) FROM PUBLIC;
