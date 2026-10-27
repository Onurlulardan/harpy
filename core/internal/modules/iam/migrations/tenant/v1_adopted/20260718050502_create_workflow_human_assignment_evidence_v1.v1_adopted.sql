-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_workflow_human_assignment_evidence_v1(
    p_scope_schema name,
    p_tenant_id uuid,
    p_assignment_kind text,
    p_assignment_id uuid
)
RETURNS TABLE (assignment_document jsonb)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_revision bigint;
    v_scope_kind text;
    v_organization_id uuid;
    v_organization_revision bigint;
    v_facts jsonb;
BEGIN
    IF p_assignment_kind IS NULL OR
       p_assignment_kind NOT IN ('user','role','organization') OR
       p_assignment_id IS NULL OR
       p_assignment_id = '00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.substr(p_assignment_id::text, 20, 1) NOT IN ('8','9','a','b') THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Workflow human assignment input is invalid';
    END IF;
    PERFORM iam_workflow_human_scope_assert_v1(p_scope_schema, p_tenant_id);
    CASE p_assignment_kind
    WHEN 'user' THEN
        SELECT actor.revision INTO STRICT v_revision
        FROM iam_users AS actor
        WHERE actor.id = p_assignment_id AND actor.identity_kind = 'human'
          AND actor.status = 'active'
        FOR SHARE;
        v_facts := pg_catalog.jsonb_build_array(pg_catalog.jsonb_build_object(
            'kind','user','id',p_assignment_id,'revision',v_revision));
    WHEN 'role' THEN
        SELECT role.revision, role.scope_kind, role.organization_id
        INTO STRICT v_revision, v_scope_kind, v_organization_id
        FROM iam_roles AS role
        WHERE role.id = p_assignment_id AND role.status = 'active'
        FOR SHARE;
        v_facts := pg_catalog.jsonb_build_array(pg_catalog.jsonb_build_object(
            'kind','role','id',p_assignment_id,'revision',v_revision));
        IF v_scope_kind = 'organization' THEN
            SELECT organization.revision INTO STRICT v_organization_revision
            FROM iam_organizations AS organization
            WHERE organization.id = v_organization_id
              AND organization.status = 'active'
            FOR SHARE;
            v_facts := v_facts || pg_catalog.jsonb_build_array(
                pg_catalog.jsonb_build_object(
                    'kind','organization','id',v_organization_id,
                    'revision',v_organization_revision));
        ELSIF v_scope_kind IS DISTINCT FROM 'tenant' OR
              v_organization_id IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = 'P5733',
                MESSAGE = 'Workflow human role scope is invalid';
        END IF;
    WHEN 'organization' THEN
        SELECT organization.revision INTO STRICT v_revision
        FROM iam_organizations AS organization
        WHERE organization.id = p_assignment_id
          AND organization.status = 'active'
        FOR SHARE;
        v_facts := pg_catalog.jsonb_build_array(pg_catalog.jsonb_build_object(
            'kind','organization','id',p_assignment_id,'revision',v_revision));
    END CASE;
    RETURN QUERY SELECT pg_catalog.jsonb_build_object(
        'kind',p_assignment_kind,'id',p_assignment_id,
        'revision',v_revision,'facts',v_facts);
EXCEPTION WHEN NO_DATA_FOUND OR TOO_MANY_ROWS THEN
    RAISE EXCEPTION USING ERRCODE = 'P5733',
        MESSAGE = 'Workflow human assignment evidence is unavailable';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_workflow_human_assignment_evidence_v1(
    name,uuid,text,uuid
) FROM PUBLIC;
