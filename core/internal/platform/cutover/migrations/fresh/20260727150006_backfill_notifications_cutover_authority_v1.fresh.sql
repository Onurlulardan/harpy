-- +goose Up
-- Migration stream version: 20260727150006.

INSERT INTO platform.cutover_groups (
    scope_kind, scope_id, group_id, state, writer, generation,
    requested_by, approved_by, reason, change_reference,
    acceptance_digest, changed_at
)
SELECT
    domain.scope_kind, domain.scope_id, 'notifications',
    domain.state, domain.writer, domain.generation,
    domain.requested_by, domain.approved_by,
    domain.reason, domain.change_reference,
    domain.acceptance_digest, domain.changed_at
FROM platform.domain_cutovers AS domain
WHERE domain.scope_kind = 'tenant'
  AND domain.domain = 'notifications'
  AND NOT EXISTS (
      SELECT 1
      FROM platform.cutover_groups AS authority
      WHERE authority.scope_kind = domain.scope_kind
        AND authority.scope_id = domain.scope_id
        AND authority.group_id = 'notifications'
  );

INSERT INTO platform.cutover_group_members (
    scope_kind, scope_id, group_id, ordinal, domain
)
SELECT authority.scope_kind, authority.scope_id, authority.group_id,
       1, 'notifications'
FROM platform.cutover_groups AS authority
WHERE authority.scope_kind = 'tenant'
  AND authority.group_id = 'notifications'
  AND NOT EXISTS (
      SELECT 1
      FROM platform.cutover_group_members AS member
      WHERE member.scope_kind = authority.scope_kind
        AND member.scope_id = authority.scope_id
        AND member.group_id = authority.group_id
        AND member.domain = 'notifications'
  );

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.cutover_initialize_tenant_v1(
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
            IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize'
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
              AND change_reference = v_reference) IS DISTINCT FROM 10
           OR (SELECT pg_catalog.count(*) FROM platform.cutover_groups
               WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id
                 AND group_id IN (
                    'billing_forms','iam_system_resources','notifications'
                 )
                 AND state = 'v2' AND writer = 'v2' AND generation = 1
                 AND requested_by = p_operation_id AND approved_by = p_owner_user_id
                 AND acceptance_digest = p_manifest_sha256
                 AND change_reference = v_reference) IS DISTINCT FROM 3
           OR (SELECT pg_catalog.count(*) FROM platform.cutover_groups
               WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id) IS DISTINCT FROM 3
           OR (SELECT pg_catalog.count(*) FROM platform.cutover_group_members
               WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id) IS DISTINCT FROM 5
           OR (SELECT pg_catalog.count(*) FROM platform.cutover_group_members
               WHERE scope_kind = 'tenant' AND scope_id = p_tenant_id
                 AND (group_id,ordinal,domain) IN (
                    ('billing_forms',1,'billing'),('billing_forms',2,'forms'),
                    ('iam_system_resources',1,'iam'),
                    ('iam_system_resources',2,'system_resource_layout'),
                    ('notifications',1,'notifications')
                 )) IS DISTINCT FROM 5 THEN
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
        'analytics','billing','catalog','documents','forms','iam','notifications',
        'system_resource_layout','work','workflow'
      ]) AS entry(domain);
    INSERT INTO platform.cutover_groups (
        scope_kind, scope_id, group_id, state, writer, generation,
        requested_by, approved_by, reason, change_reference, acceptance_digest
    ) SELECT 'tenant', p_tenant_id, group_id, 'v2', 'v2', 1,
        p_operation_id, p_owner_user_id, 'fresh V2 tenant activation',
        v_reference, p_manifest_sha256
      FROM pg_catalog.unnest(ARRAY[
        'billing_forms','iam_system_resources','notifications'
      ]) AS entry(group_id);
    INSERT INTO platform.cutover_group_members (
        scope_kind, scope_id, group_id, ordinal, domain
    ) VALUES
        ('tenant',p_tenant_id,'billing_forms',1,'billing'),
        ('tenant',p_tenant_id,'billing_forms',2,'forms'),
        ('tenant',p_tenant_id,'iam_system_resources',1,'iam'),
        ('tenant',p_tenant_id,'iam_system_resources',2,'system_resource_layout'),
        ('tenant',p_tenant_id,'notifications',1,'notifications');
END;
$function$;
-- +goose StatementEnd

SELECT 1 / CASE WHEN EXISTS (
        SELECT 1
        FROM platform.domain_cutovers AS domain
        LEFT JOIN platform.cutover_groups AS authority
          ON authority.scope_kind = domain.scope_kind
         AND authority.scope_id = domain.scope_id
         AND authority.group_id = 'notifications'
        LEFT JOIN platform.cutover_group_members AS member
          ON member.scope_kind = authority.scope_kind
         AND member.scope_id = authority.scope_id
         AND member.group_id = authority.group_id
         AND member.ordinal = 1
         AND member.domain = 'notifications'
        WHERE domain.scope_kind = 'tenant'
          AND domain.domain = 'notifications'
          AND (
              authority.scope_id IS NULL
              OR authority.state IS DISTINCT FROM domain.state
              OR authority.writer IS DISTINCT FROM domain.writer
              OR authority.generation IS DISTINCT FROM domain.generation
              OR member.scope_id IS NULL
          )
    ) THEN 0 ELSE 1 END;
