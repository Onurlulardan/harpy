-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_workflow_human_candidate_evidence_v1(
    p_scope_schema name,
    p_tenant_id uuid,
    p_actor_id uuid,
    p_candidate_kind text,
    p_candidate_id uuid
)
RETURNS TABLE (evidence_document jsonb)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_authorization_version bigint;
    v_user_revision bigint;
    v_role_revision bigint;
    v_role_scope text;
    v_organization_id uuid;
    v_organization_revision bigint;
    v_assignment_id uuid;
    v_assignment_revision bigint;
    v_assignment_scope text;
    v_assignment_organization uuid;
    v_membership_id uuid;
    v_membership_revision bigint;
    v_facts jsonb;
BEGIN
    IF p_actor_id IS NULL OR p_candidate_id IS NULL OR
       p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       p_candidate_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       p_candidate_kind IS NULL OR
       p_candidate_kind NOT IN ('user','role','organization') OR
       pg_catalog.substr(p_actor_id::text, 20, 1) NOT IN ('8','9','a','b') OR
       pg_catalog.substr(p_candidate_id::text, 20, 1) NOT IN ('8','9','a','b') THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Workflow human candidate input is invalid';
    END IF;
    PERFORM iam_workflow_human_scope_assert_v1(p_scope_schema, p_tenant_id);
    SELECT actor.authorization_version, actor.revision
    INTO STRICT v_authorization_version, v_user_revision
    FROM iam_users AS actor
    WHERE actor.id = p_actor_id AND actor.identity_kind = 'human'
      AND actor.status = 'active' AND actor.authorization_version > 0
    FOR SHARE;
    CASE p_candidate_kind
    WHEN 'user' THEN
        IF p_candidate_id IS DISTINCT FROM p_actor_id THEN
            RAISE EXCEPTION USING ERRCODE = 'P5734',
                MESSAGE = 'Workflow human user candidate does not match actor';
        END IF;
        v_facts := pg_catalog.jsonb_build_array(pg_catalog.jsonb_build_object(
            'kind','user','id',p_actor_id,'revision',v_user_revision));
    WHEN 'role' THEN
        SELECT role.revision, role.scope_kind, role.organization_id
        INTO STRICT v_role_revision, v_role_scope, v_organization_id
        FROM iam_roles AS role
        WHERE role.id = p_candidate_id AND role.status = 'active'
        FOR SHARE;
        SELECT assignment.id, assignment.revision, assignment.role_scope_kind,
               assignment.organization_id, assignment.membership_id
        INTO STRICT v_assignment_id, v_assignment_revision, v_assignment_scope,
                    v_assignment_organization, v_membership_id
        FROM iam_user_role_assignments AS assignment
        WHERE assignment.user_id = p_actor_id
          AND assignment.role_id = p_candidate_id AND assignment.state = 'active'
        FOR SHARE;
        IF v_assignment_scope IS DISTINCT FROM v_role_scope OR
           v_assignment_organization IS DISTINCT FROM v_organization_id THEN
            RAISE EXCEPTION USING ERRCODE = 'P5734',
                MESSAGE = 'Workflow human role assignment scope is invalid';
        END IF;
        v_facts := pg_catalog.jsonb_build_array(
            pg_catalog.jsonb_build_object(
                'kind','role','id',p_candidate_id,'revision',v_role_revision),
            pg_catalog.jsonb_build_object(
                'kind','role_assignment','id',v_assignment_id,
                'revision',v_assignment_revision));
        IF v_role_scope = 'organization' THEN
            SELECT organization.revision INTO STRICT v_organization_revision
            FROM iam_organizations AS organization
            WHERE organization.id = v_organization_id
              AND organization.status = 'active'
            FOR SHARE;
            SELECT membership.revision INTO STRICT v_membership_revision
            FROM iam_organization_memberships AS membership
            WHERE membership.id = v_membership_id
              AND membership.organization_id = v_organization_id
              AND membership.user_id = p_actor_id AND membership.state = 'active'
            FOR SHARE;
            v_facts := v_facts || pg_catalog.jsonb_build_array(
                pg_catalog.jsonb_build_object(
                    'kind','organization','id',v_organization_id,
                    'revision',v_organization_revision),
                pg_catalog.jsonb_build_object(
                    'kind','organization_membership','id',v_membership_id,
                    'revision',v_membership_revision));
        ELSIF v_role_scope IS DISTINCT FROM 'tenant' OR
              v_organization_id IS NOT NULL OR v_membership_id IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = 'P5734',
                MESSAGE = 'Workflow human role candidate scope is invalid';
        END IF;
    WHEN 'organization' THEN
        SELECT organization.revision INTO STRICT v_organization_revision
        FROM iam_organizations AS organization
        WHERE organization.id = p_candidate_id
          AND organization.status = 'active'
        FOR SHARE;
        SELECT membership.id, membership.revision
        INTO STRICT v_membership_id, v_membership_revision
        FROM iam_organization_memberships AS membership
        WHERE membership.organization_id = p_candidate_id
          AND membership.user_id = p_actor_id AND membership.state = 'active'
        FOR SHARE;
        v_facts := pg_catalog.jsonb_build_array(
            pg_catalog.jsonb_build_object(
                'kind','organization','id',p_candidate_id,
                'revision',v_organization_revision),
            pg_catalog.jsonb_build_object(
                'kind','organization_membership','id',v_membership_id,
                'revision',v_membership_revision));
    END CASE;
    RETURN QUERY SELECT pg_catalog.jsonb_build_object(
        'actor_id',p_actor_id,'authorization_version',v_authorization_version,
        'candidate_kind',p_candidate_kind,'candidate_id',p_candidate_id,
        'facts',v_facts);
EXCEPTION WHEN NO_DATA_FOUND OR TOO_MANY_ROWS THEN
    RAISE EXCEPTION USING ERRCODE = 'P5734',
        MESSAGE = 'Workflow human candidate evidence is unavailable';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_workflow_human_candidate_evidence_v1(
    name,uuid,uuid,text,uuid
) FROM PUBLIC;
