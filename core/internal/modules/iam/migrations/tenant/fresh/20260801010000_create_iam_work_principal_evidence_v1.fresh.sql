-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_work_principal_evidence_v1(
    p_scope_schema name,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_target_kind text,
    p_target_id uuid
)
RETURNS TABLE (
    actor_principals jsonb,
    authorization_clock bigint,
    target_exists boolean,
    target_revision bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_iam_generation bigint;
    v_work_generation bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS NULL OR p_actor_id IS NULL OR p_target_id IS NULL OR
       p_target_kind NOT IN ('USER', 'ROLE', 'ORGANIZATION') OR
       NOT EXISTS (
           SELECT 1 FROM tenancy_schema_identity AS identity_row
           WHERE identity_row.singleton AND identity_row.tenant_id = p_tenant_id
             AND identity_row.schema_name = p_scope_schema::text
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM Work principal evidence scope is invalid';
    END IF;

    SELECT authority.generation INTO v_iam_generation
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind = 'tenant' AND authority.scope_id = p_tenant_id
      AND authority.domain = 'iam'
      AND authority.state = 'v2' AND authority.writer = 'v2';
    SELECT authority.generation INTO v_work_generation
    FROM platform.domain_cutovers AS authority
    WHERE authority.scope_kind = 'tenant' AND authority.scope_id = p_tenant_id
      AND authority.domain = 'work'
      AND authority.state = 'v2' AND authority.writer = 'v2';
    IF v_iam_generation IS NULL OR v_iam_generation < 1 OR
       v_work_generation IS NULL OR v_work_generation < 1 OR
       NOT EXISTS (
           SELECT 1 FROM platform.cutover_group_members AS member
           WHERE member.scope_kind = 'tenant' AND member.scope_id = p_tenant_id
             AND member.group_id = 'work' AND member.ordinal = 1
             AND member.domain = 'work'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5701',
            MESSAGE = 'IAM Work principal evidence authority is unavailable';
    END IF;

    RETURN QUERY
    WITH actor AS (
        SELECT principal.id
        FROM iam_users AS principal
        WHERE principal.id = p_actor_id AND principal.status = 'active'
          AND EXISTS (
              SELECT 1
              FROM iam_user_access_modes AS access_mode
              WHERE access_mode.user_id = principal.id
                AND access_mode.mode = 'operator_console'
                AND access_mode.state = 'active'
          )
    ), principals AS (
        SELECT 'USER'::text AS kind, actor.id, user_row.revision
        FROM actor JOIN iam_users AS user_row ON user_row.id = actor.id
        UNION ALL
        SELECT 'ROLE', role.id,
            GREATEST(role.revision, assignment.revision,
                COALESCE(membership.revision, 0), COALESCE(organization.revision, 0))
        FROM actor
        JOIN iam_user_role_assignments AS assignment
          ON assignment.user_id = actor.id AND assignment.state = 'active'
        JOIN iam_roles AS role ON role.id = assignment.role_id AND role.status = 'active'
        LEFT JOIN iam_organization_memberships AS membership
          ON membership.id = assignment.membership_id AND membership.state = 'active'
        LEFT JOIN iam_organizations AS organization
          ON organization.id = assignment.organization_id AND organization.status = 'active'
        WHERE assignment.role_scope_kind = 'tenant' OR
              (membership.id IS NOT NULL AND organization.id IS NOT NULL)
        UNION ALL
        SELECT 'ORGANIZATION', organization.id,
            GREATEST(organization.revision, membership.revision)
        FROM actor
        JOIN iam_organization_memberships AS membership
          ON membership.user_id = actor.id AND membership.state = 'active'
        JOIN iam_organizations AS organization
          ON organization.id = membership.organization_id
         AND organization.status = 'active'
    ), target AS (
        SELECT user_row.revision
        FROM iam_users AS user_row
        WHERE p_target_kind = 'USER' AND user_row.id = p_target_id
          AND user_row.status = 'active'
        UNION ALL
        SELECT role.revision FROM iam_roles AS role
        WHERE p_target_kind = 'ROLE' AND role.id = p_target_id
          AND role.status = 'active'
        UNION ALL
        SELECT organization.revision FROM iam_organizations AS organization
        WHERE p_target_kind = 'ORGANIZATION' AND organization.id = p_target_id
          AND organization.status = 'active'
    )
    SELECT COALESCE((
        SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
            'kind', kind, 'id', id, 'revision', revision
        ) ORDER BY kind, id) FROM principals
    ), '[]'::jsonb),
    COALESCE((SELECT clock_value FROM iam_authorization_clock WHERE singleton), -1),
    EXISTS (SELECT 1 FROM target),
    COALESCE((SELECT revision FROM target), -1);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_work_principal_evidence_v1(name,uuid,uuid,text,uuid)
    FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_work_principal_evidence_v1(name,uuid,uuid,text,uuid)
    TO forgecrud_runtime;
