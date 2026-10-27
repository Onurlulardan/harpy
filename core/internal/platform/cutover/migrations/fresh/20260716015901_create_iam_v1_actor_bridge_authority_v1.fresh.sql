-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_v1_actor_bridge_authority_v1(
    p_tenant_id uuid,
    p_target_domain text
)
RETURNS TABLE (
    iam_group_generation bigint,
    iam_generation bigint,
    target_generation bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_group platform.cutover_groups%ROWTYPE;
    v_iam platform.domain_cutovers%ROWTYPE;
    v_target platform.domain_cutovers%ROWTYPE;
    v_member_count bigint;
    v_exact_member_count bigint;
    v_retirement platform.contract_retirements%ROWTYPE;
BEGIN
    IF SESSION_USER NOT IN ('forgecrud_v1_runtime','forgecrud_v2_runtime') OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_tenant_id IS NULL OR p_target_domain IS NULL OR
       NOT (p_target_domain = ANY (ARRAY[
           'analytics','billing','catalog','documents','forms',
           'notifications','work','workflow'
       ]::text[])) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM V1 actor bridge authority input is invalid';
    END IF;

    SELECT authority_group.* INTO v_group
    FROM platform.cutover_groups AS authority_group
    JOIN platform.cutover_group_members AS member
      ON member.scope_kind = authority_group.scope_kind
     AND member.scope_id = authority_group.scope_id
     AND member.group_id = authority_group.group_id
    WHERE authority_group.scope_kind = 'tenant'
      AND authority_group.scope_id = p_tenant_id
      AND member.domain = 'iam'
    FOR SHARE OF authority_group, member;
    IF NOT FOUND OR v_group.group_id IS DISTINCT FROM 'iam_system_resources' OR
       v_group.state IS DISTINCT FROM 'v2' OR v_group.writer IS DISTINCT FROM 'v2' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5601',
            MESSAGE = 'IAM V1 actor bridge IAM group is not V2';
    END IF;

    PERFORM member.ordinal
    FROM platform.cutover_group_members AS member
    WHERE member.scope_kind = 'tenant' AND member.scope_id = p_tenant_id
      AND member.group_id = 'iam_system_resources'
    ORDER BY member.ordinal
    FOR SHARE;
    SELECT pg_catalog.count(*),
           pg_catalog.count(*) FILTER (
               WHERE member.domain IN ('iam','system_resource_layout')
           )
    INTO v_member_count, v_exact_member_count
    FROM platform.cutover_group_members AS member
    WHERE member.scope_kind = 'tenant' AND member.scope_id = p_tenant_id
      AND member.group_id = 'iam_system_resources';
    IF v_member_count IS DISTINCT FROM 2 OR v_exact_member_count IS DISTINCT FROM 2 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5601',
            MESSAGE = 'IAM V1 actor bridge IAM group membership differs';
    END IF;

    PERFORM authority.domain
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind = 'tenant' AND authority.scope_id = p_tenant_id
      AND authority.domain IN ('iam', p_target_domain)
    ORDER BY authority.domain COLLATE "C"
    FOR SHARE;
    SELECT authority.* INTO v_iam
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind = 'tenant' AND authority.scope_id = p_tenant_id
      AND authority.domain = 'iam';
    SELECT authority.* INTO v_target
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind = 'tenant' AND authority.scope_id = p_tenant_id
      AND authority.domain = p_target_domain;
    IF v_iam.state IS DISTINCT FROM 'v2' OR v_iam.writer IS DISTINCT FROM 'v2' OR
       v_target.state NOT IN ('v1','shadow') OR v_target.writer IS DISTINCT FROM 'v1' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5602',
            MESSAGE = 'IAM V1 actor bridge domain authority differs';
    END IF;

    SELECT retirement.* INTO v_retirement
    FROM platform.contract_retirements AS retirement
    WHERE retirement.contract_key = 'iam.v1_actor_assertion.v1';
    IF NOT FOUND OR v_retirement.owner_module IS DISTINCT FROM 'iam' OR
       v_retirement.domain IS DISTINCT FROM 'iam' OR
       v_retirement.state IS DISTINCT FROM 'active' OR
       v_retirement.remaining_scope_count < 1 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5603',
            MESSAGE = 'IAM V1 actor bridge contract is unavailable';
    END IF;

    RETURN QUERY SELECT v_group.generation, v_iam.generation, v_target.generation;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_v1_actor_bridge_authority_v1(uuid,text)
FROM PUBLIC;
