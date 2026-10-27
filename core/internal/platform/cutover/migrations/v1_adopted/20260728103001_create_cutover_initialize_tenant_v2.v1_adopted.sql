-- +goose Up
-- Migration stream version: 20260728103001.
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_initialize_tenant_v2(
    p_operation_id uuid,
    p_tenant_id uuid,
    p_owner_user_id uuid,
    p_manifest_sha256 text
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_reference text;
BEGIN
    v_reference := 'installation:' || p_operation_id::text;

    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize'
       OR p_operation_id IS NULL OR p_tenant_id IS NULL
       OR p_owner_user_id IS NULL OR p_operation_id = p_owner_user_id
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant cutover initialization v2 denied';
    END IF;
    IF (SELECT pg_catalog.count(*) FROM platform.domain_cutovers
        WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id) <> 10
       OR (SELECT pg_catalog.count(*) FROM platform.domain_cutovers
           WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id
             AND domain IN (
                 'analytics','billing','catalog','documents','forms','iam',
                 'notifications','system_resource_layout','work','workflow'
             )
             AND state = 'v2' AND writer = 'v2' AND generation = 1
             AND requested_by = p_operation_id
             AND approved_by = p_owner_user_id
             AND acceptance_digest = p_manifest_sha256
             AND change_reference = v_reference) <> 10 THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'tenant cutover initialization v2 domain conflict';
    END IF;
    INSERT INTO platform.cutover_groups (
        scope_kind, scope_id, group_id, state, writer, generation,
        requested_by, approved_by, reason, change_reference,
        acceptance_digest, changed_at
    )
    SELECT domain.scope_kind, domain.scope_id, domain.domain,
           domain.state, domain.writer, domain.generation,
           domain.requested_by, domain.approved_by, domain.reason,
           domain.change_reference, domain.acceptance_digest, domain.changed_at
    FROM platform.domain_cutovers AS domain
    WHERE domain.scope_kind = 'tenant' AND domain.scope_id = p_tenant_id
      AND domain.domain IN (
          'analytics', 'catalog', 'documents',
          'notifications', 'work', 'workflow'
      )
    ON CONFLICT (scope_kind, scope_id, group_id) DO NOTHING;

    INSERT INTO platform.cutover_group_members (
        scope_kind, scope_id, group_id, ordinal, domain
    )
    SELECT authority.scope_kind, authority.scope_id,
           authority.group_id, 1, authority.group_id
    FROM platform.cutover_groups AS authority
    WHERE authority.scope_kind = 'tenant'
      AND authority.scope_id = p_tenant_id
      AND authority.group_id IN (
          'analytics', 'catalog', 'documents',
          'notifications', 'work', 'workflow'
      )
    ON CONFLICT (scope_kind, scope_id, group_id, ordinal) DO NOTHING;

    IF (SELECT pg_catalog.count(*) FROM platform.cutover_groups
        WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id) <> 8
       OR (SELECT pg_catalog.count(*) FROM platform.cutover_groups
           WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id
             AND group_id IN (
                 'analytics','billing_forms','catalog','documents',
                 'iam_system_resources','notifications','work','workflow'
             )
             AND state = 'v2' AND writer = 'v2' AND generation = 1
             AND requested_by = p_operation_id
             AND approved_by = p_owner_user_id
             AND acceptance_digest = p_manifest_sha256
             AND change_reference = v_reference) <> 8
       OR (SELECT pg_catalog.count(*) FROM platform.cutover_group_members
           WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id) <> 10
       OR (SELECT pg_catalog.count(*) FROM platform.cutover_group_members
           WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id
             AND (group_id,ordinal,domain) IN (
                 ('analytics',1,'analytics'),
                 ('billing_forms',1,'billing'),('billing_forms',2,'forms'),
                 ('catalog',1,'catalog'),('documents',1,'documents'),
                 ('iam_system_resources',1,'iam'),
                 ('iam_system_resources',2,'system_resource_layout'),
                 ('notifications',1,'notifications'),('work',1,'work'),
                 ('workflow',1,'workflow')
             )) <> 10 THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'tenant cutover initialization v2 authority conflict';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.cutover_initialize_tenant_v2(
    uuid,uuid,uuid,text
) FROM PUBLIC;
