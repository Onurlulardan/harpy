-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_workflow_human_scope_assert_v1(
    p_scope_schema name,
    p_tenant_id uuid
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_group_generation bigint;
    v_iam_generation bigint;
    v_layout_generation bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS NULL OR
       p_tenant_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       NOT EXISTS (
           SELECT 1 FROM tenancy_schema_identity AS identity_row
           WHERE identity_row.singleton
             AND identity_row.tenant_id = p_tenant_id
             AND identity_row.schema_name = p_scope_schema::text
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Workflow human IAM evidence scope is invalid';
    END IF;
    SELECT authority.generation INTO v_group_generation
    FROM platform.cutover_groups AS authority
    WHERE authority.scope_kind = 'tenant'
      AND authority.scope_id = p_tenant_id
      AND authority.group_id = 'iam_system_resources'
      AND authority.state = 'v2' AND authority.writer = 'v2'
    FOR SHARE;
    IF NOT FOUND OR v_group_generation < 1 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5731',
            MESSAGE = 'Workflow human IAM group authority is not current';
    END IF;
    SELECT authority.generation INTO v_iam_generation
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind = 'tenant'
      AND authority.scope_id = p_tenant_id AND authority.domain = 'iam'
      AND authority.state = 'v2' AND authority.writer = 'v2'
    FOR SHARE;
    SELECT authority.generation INTO v_layout_generation
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind = 'tenant'
      AND authority.scope_id = p_tenant_id
      AND authority.domain = 'system_resource_layout'
      AND authority.state = 'v2' AND authority.writer = 'v2'
    FOR SHARE;
    IF v_iam_generation IS DISTINCT FROM v_group_generation OR
       v_layout_generation IS DISTINCT FROM v_group_generation OR
       (SELECT pg_catalog.count(*) FROM platform.cutover_group_members AS member
        WHERE member.scope_kind = 'tenant' AND member.scope_id = p_tenant_id
          AND member.group_id = 'iam_system_resources') <> 2 OR
       NOT EXISTS (
           SELECT 1 FROM platform.cutover_group_members AS member
           WHERE member.scope_kind = 'tenant' AND member.scope_id = p_tenant_id
             AND member.group_id = 'iam_system_resources'
             AND member.ordinal = 1 AND member.domain = 'iam'
       ) OR NOT EXISTS (
           SELECT 1 FROM platform.cutover_group_members AS member
           WHERE member.scope_kind = 'tenant' AND member.scope_id = p_tenant_id
             AND member.group_id = 'iam_system_resources'
             AND member.ordinal = 2 AND member.domain = 'system_resource_layout'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5731',
            MESSAGE = 'Workflow human IAM domain authority is not current';
    END IF;
END;
$function$;

REVOKE ALL ON FUNCTION iam_workflow_human_scope_assert_v1(name,uuid)
    FROM PUBLIC;

CREATE FUNCTION iam_authorization_evidence_read_v4(
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
BEGIN
    IF p_actor_id IS NULL OR
       p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.substr(p_actor_id::text, 20, 1) NOT IN ('8','9','a','b') OR
       p_resource_key IS DISTINCT FROM 'workflow.human_tasks' OR
       p_action_key NOT IN ('claim','decide','reassign','expire','cancel') OR
       pg_catalog.char_length(p_resource_key) NOT BETWEEN 1 AND 100 OR
       pg_catalog.char_length(p_action_key) NOT BETWEEN 1 AND 100 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Workflow human authorization input is invalid';
    END IF;
    PERFORM iam_workflow_human_scope_assert_v1(p_scope_schema, p_tenant_id);
    PERFORM 1 FROM iam_users AS actor
    WHERE actor.id = p_actor_id AND actor.identity_kind = 'human'
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5732',
            MESSAGE = 'Workflow human actor is unavailable';
    END IF;
    RETURN QUERY
    SELECT evidence.actor_document, evidence.resource_document,
           evidence.action_document, evidence.authorization_clock,
           evidence.bindings_document, evidence.implications_document
    FROM iam_authorization_evidence_read_v2(
        p_scope_schema, p_tenant_id, p_actor_id, p_resource_key, p_action_key
    ) AS evidence;
END;
$function$;

REVOKE ALL ON FUNCTION iam_authorization_evidence_read_v4(
    name,uuid,uuid,text,text
) FROM PUBLIC;
-- +goose StatementEnd
