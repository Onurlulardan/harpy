-- +goose Up
-- Migration stream version: 20260722080300.
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_initialize_tenant_v1(
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
    v_existing integer;
BEGIN
    v_reference := 'installation:' || p_operation_id::text;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'harpy-v2:core-operator-initialize'
       OR p_operation_id IS NULL OR p_tenant_id IS NULL
       OR p_owner_user_id IS NULL OR p_operation_id = p_owner_user_id
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant cutover initialization denied';
    END IF;
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
    IF v_existing > 0 THEN
        IF (SELECT pg_catalog.count(*) FROM platform.domain_cutovers
            WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id
              AND state = 'v2' AND writer = 'v2' AND generation = 1
              AND requested_by = p_operation_id AND approved_by = p_owner_user_id
              AND acceptance_digest = p_manifest_sha256
              AND change_reference = v_reference) IS DISTINCT FROM 9
           OR (SELECT pg_catalog.count(*) FROM platform.cutover_groups
               WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id
                 AND group_id IN ('billing_forms','iam_system_resources')
                 AND state = 'v2' AND writer = 'v2' AND generation = 1
                 AND requested_by = p_operation_id AND approved_by = p_owner_user_id
                 AND acceptance_digest = p_manifest_sha256
                 AND change_reference = v_reference) IS DISTINCT FROM 2
           OR (SELECT pg_catalog.count(*) FROM platform.cutover_groups
               WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id) IS DISTINCT FROM 2
           OR (SELECT pg_catalog.count(*) FROM platform.cutover_group_members
               WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id) IS DISTINCT FROM 3
           OR (SELECT pg_catalog.count(*) FROM platform.cutover_group_members
               WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id
                 AND (group_id,ordinal,domain) IN (
                    ('billing_forms',1,'billing'),
                    ('iam_system_resources',1,'iam'),
                    ('iam_system_resources',2,'system_resource_layout')
                 )) IS DISTINCT FROM 3 THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'tenant cutover initialization replay conflicts';
        END IF;
        RETURN;
    END IF;
    INSERT INTO platform.domain_cutovers (
        scope_kind, scope_id, domain, state, writer, generation,
        requested_by, approved_by, reason, change_reference, acceptance_digest
    ) SELECT 'tenant', p_tenant_id, domain, 'v2', 'v2', 1,
        p_operation_id, p_owner_user_id, 'fresh V2 tenant activation',
        v_reference, p_manifest_sha256
      FROM pg_catalog.unnest(ARRAY[
        'analytics','billing','catalog','documents','iam','notifications',
        'system_resource_layout','work','workflow'
      ]) AS entry(domain);
    INSERT INTO platform.cutover_groups (
        scope_kind, scope_id, group_id, state, writer, generation,
        requested_by, approved_by, reason, change_reference, acceptance_digest
    ) SELECT 'tenant', p_tenant_id, group_id, 'v2', 'v2', 1,
        p_operation_id, p_owner_user_id, 'fresh V2 tenant activation',
        v_reference, p_manifest_sha256
      FROM pg_catalog.unnest(
        ARRAY['billing_forms','iam_system_resources']
      ) AS entry(group_id);
    INSERT INTO platform.cutover_group_members (
        scope_kind, scope_id, group_id, ordinal, domain
    ) VALUES
        ('tenant',p_tenant_id,'billing_forms',1,'billing'),
        ('tenant',p_tenant_id,'iam_system_resources',1,'iam'),
        ('tenant',p_tenant_id,'iam_system_resources',2,'system_resource_layout');
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.cutover_initialize_tenant_v1(
    uuid,uuid,uuid,text
) FROM PUBLIC;
