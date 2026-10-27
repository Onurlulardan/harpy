-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_authorization_evidence_read_v2(
    p_scope_schema name,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_resource_key text,
    p_action_key text
)
RETURNS TABLE (
    actor_document jsonb,
    resource_document jsonb,
    action_document jsonb,
    authorization_clock bigint,
    bindings_document jsonb,
    implications_document jsonb
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_group_generation bigint;
    v_iam_generation bigint;
    v_layout_generation bigint;
    v_clock bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS NULL OR p_actor_id IS NULL OR
       p_resource_key IS NULL OR p_action_key IS NULL OR
       pg_catalog.char_length(p_resource_key) NOT BETWEEN 1 AND 100 OR
       pg_catalog.char_length(p_action_key) NOT BETWEEN 1 AND 100 OR
       NOT EXISTS (
           SELECT 1
           FROM tenancy_schema_identity AS identity_row
           WHERE identity_row.singleton
             AND identity_row.tenant_id = p_tenant_id
             AND identity_row.schema_name = p_scope_schema::text
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM authorization evidence V2 scope is invalid';
    END IF;

    SELECT authority.generation
    INTO v_group_generation
    FROM platform.cutover_groups AS authority
    WHERE authority.scope_kind = 'tenant'
      AND authority.scope_id = p_tenant_id
      AND authority.group_id = 'iam_system_resources'
      AND authority.state = 'v2'
      AND authority.writer = 'v2'
    FOR SHARE;
    IF NOT FOUND OR v_group_generation < 1 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5701',
            MESSAGE = 'IAM authorization authority is not current V2';
    END IF;
    SELECT authority.generation
    INTO v_iam_generation
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind = 'tenant'
      AND authority.scope_id = p_tenant_id
      AND authority.domain = 'iam'
      AND authority.state = 'v2'
      AND authority.writer = 'v2'
    FOR SHARE;
    IF NOT FOUND OR v_iam_generation IS DISTINCT FROM v_group_generation THEN
        RAISE EXCEPTION USING ERRCODE = 'P5702',
            MESSAGE = 'IAM domain authority is not current V2';
    END IF;

    SELECT authority.generation
    INTO v_layout_generation
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind = 'tenant'
      AND authority.scope_id = p_tenant_id
      AND authority.domain = 'system_resource_layout'
      AND authority.state = 'v2'
      AND authority.writer = 'v2'
    FOR SHARE;
    IF NOT FOUND OR v_layout_generation IS DISTINCT FROM v_group_generation THEN
        RAISE EXCEPTION USING ERRCODE = 'P5703',
            MESSAGE = 'IAM layout domain authority is not current V2';
    END IF;

    PERFORM 1
    FROM platform.cutover_group_members AS member
    WHERE member.scope_kind = 'tenant'
      AND member.scope_id = p_tenant_id
      AND member.group_id = 'iam_system_resources'
      AND member.ordinal = 1
      AND member.domain = 'iam'
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5704',
            MESSAGE = 'IAM group membership is invalid';
    END IF;
    PERFORM 1
    FROM platform.cutover_group_members AS member
    WHERE member.scope_kind = 'tenant'
      AND member.scope_id = p_tenant_id
      AND member.group_id = 'iam_system_resources'
      AND member.ordinal = 2
      AND member.domain = 'system_resource_layout'
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5704',
            MESSAGE = 'IAM group membership is invalid';
    END IF;

    PERFORM 1 FROM iam_users WHERE id = p_actor_id FOR SHARE;
    SELECT clock_value INTO v_clock
    FROM iam_authorization_clock
    WHERE singleton
    FOR SHARE;
    IF NOT FOUND OR v_clock < 0 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5705',
            MESSAGE = 'IAM authorization clock is unavailable';
    END IF;

    RETURN QUERY
    SELECT evidence.actor_document, evidence.resource_document,
           evidence.action_document, evidence.authorization_clock,
           evidence.bindings_document, evidence.implications_document
    FROM iam_authorization_evidence_read_v1(
        p_scope_schema, p_actor_id, p_resource_key, p_action_key
    ) AS evidence;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_authorization_evidence_read_v2(
    name,uuid,uuid,text,text
) FROM PUBLIC;
