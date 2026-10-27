-- +goose Up
-- Migration stream version: 20260802235003.
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_initialize_tenant_v4(
    p_operation_id uuid,
    p_tenant_id uuid,
    p_owner_user_id uuid,
    p_manifest_sha256 text,
    p_operation_kind text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_application text;
    v_reference text;
    v_existing integer;
BEGIN
    v_application := pg_catalog.current_setting('application_name', true);
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR NOT (
           (p_operation_kind = 'environment_initialize'
            AND v_application = 'harpy-v2:core-operator-initialize')
           OR (p_operation_kind = 'saas_customer_tenant_provision'
               AND v_application = 'harpy-v2:core-worker')
       )
       OR p_operation_id IS NULL OR p_tenant_id IS NULL
       OR p_owner_user_id IS NULL OR p_operation_id = p_owner_user_id
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant cutover initialization v4 denied';
    END IF;
    IF p_operation_kind = 'environment_initialize' THEN
        PERFORM platform.cutover_initialize_tenant_v2(
            p_operation_id, p_tenant_id, p_owner_user_id, p_manifest_sha256
        );
        PERFORM platform.cutover_ensure_mcp_ai_authority_v1(p_tenant_id);
        RETURN;
    END IF;
    v_reference := 'customer-activation:' || p_operation_id::text;
    SELECT pg_catalog.count(*) INTO v_existing
    FROM (
        SELECT 1 FROM platform.domain_cutovers
        WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id
        UNION ALL
        SELECT 1 FROM platform.cutover_groups
        WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id
        UNION ALL
        SELECT 1 FROM platform.cutover_group_members
        WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id
    ) AS existing;
    IF v_existing = 0 THEN
        INSERT INTO platform.domain_cutovers (
            scope_kind, scope_id, domain, state, writer, generation,
            requested_by, approved_by, reason, change_reference,
            acceptance_digest
        )
        SELECT 'tenant', p_tenant_id, domain, 'v2', 'v2', 1,
            p_operation_id, p_owner_user_id, 'SaaS customer V2 activation',
            v_reference, p_manifest_sha256
        FROM pg_catalog.unnest(ARRAY[
            'ai','analytics','billing','catalog','documents','forms','iam',
            'mcp','notifications','system_resource_layout','work','workflow'
        ]) AS entry(domain);
        INSERT INTO platform.cutover_groups (
            scope_kind, scope_id, group_id, state, writer, generation,
            requested_by, approved_by, reason, change_reference,
            acceptance_digest
        )
        SELECT 'tenant', p_tenant_id, group_id, 'v2', 'v2', 1,
            p_operation_id, p_owner_user_id, 'SaaS customer V2 activation',
            v_reference, p_manifest_sha256
        FROM pg_catalog.unnest(ARRAY[
            'ai','analytics','billing_forms','catalog','documents',
            'iam_system_resources','mcp','notifications','work','workflow'
        ]) AS entry(group_id);
        INSERT INTO platform.cutover_group_members (
            scope_kind, scope_id, group_id, ordinal, domain
        ) VALUES
            ('tenant',p_tenant_id,'ai',1,'ai'),
            ('tenant',p_tenant_id,'analytics',1,'analytics'),
            ('tenant',p_tenant_id,'billing_forms',1,'billing'),
            ('tenant',p_tenant_id,'billing_forms',2,'forms'),
            ('tenant',p_tenant_id,'catalog',1,'catalog'),
            ('tenant',p_tenant_id,'documents',1,'documents'),
            ('tenant',p_tenant_id,'iam_system_resources',1,'iam'),
            ('tenant',p_tenant_id,'iam_system_resources',2,'system_resource_layout'),
            ('tenant',p_tenant_id,'mcp',1,'mcp'),
            ('tenant',p_tenant_id,'notifications',1,'notifications'),
            ('tenant',p_tenant_id,'work',1,'work'),
            ('tenant',p_tenant_id,'workflow',1,'workflow');
    END IF;
    PERFORM platform.cutover_ensure_mcp_ai_authority_v1(p_tenant_id);
    IF (SELECT pg_catalog.count(*) FROM platform.domain_cutovers
        WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id
          AND domain IN (
              'ai','analytics','billing','catalog','documents','forms','iam',
              'mcp','notifications','system_resource_layout','work','workflow'
          )
          AND state = 'v2' AND writer = 'v2' AND generation = 1
          AND requested_by = p_operation_id AND approved_by = p_owner_user_id
          AND acceptance_digest = p_manifest_sha256
          AND change_reference = v_reference) <> 12
       OR (SELECT pg_catalog.count(*) FROM platform.domain_cutovers
           WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id) <> 12
       OR (SELECT pg_catalog.count(*) FROM platform.cutover_groups
           WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id
             AND group_id IN (
                 'ai','analytics','billing_forms','catalog','documents',
                 'iam_system_resources','mcp','notifications','work','workflow'
             )
             AND state = 'v2' AND writer = 'v2' AND generation = 1
             AND requested_by = p_operation_id AND approved_by = p_owner_user_id
             AND acceptance_digest = p_manifest_sha256
             AND change_reference = v_reference) <> 10
       OR (SELECT pg_catalog.count(*) FROM platform.cutover_groups
           WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id) <> 10
       OR (SELECT pg_catalog.count(*) FROM platform.cutover_group_members
           WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id) <> 12
       OR (SELECT pg_catalog.count(*) FROM platform.cutover_group_members
           WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id
             AND (group_id,ordinal,domain) IN (
                 ('ai',1,'ai'),('analytics',1,'analytics'),
                 ('billing_forms',1,'billing'),('billing_forms',2,'forms'),
                 ('catalog',1,'catalog'),('documents',1,'documents'),
                 ('iam_system_resources',1,'iam'),
                 ('iam_system_resources',2,'system_resource_layout'),
                 ('mcp',1,'mcp'),('notifications',1,'notifications'),
                 ('work',1,'work'),('workflow',1,'workflow')
             )) <> 12 THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'tenant cutover initialization v4 authority conflict';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.cutover_initialize_tenant_v4(
    uuid,uuid,uuid,text,text
) FROM PUBLIC;
