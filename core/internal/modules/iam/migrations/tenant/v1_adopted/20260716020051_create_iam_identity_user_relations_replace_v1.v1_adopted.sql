-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_user_relations_replace_v1(
    p_scope_schema name, p_operation_id uuid, p_command_key text, p_request_sha256 bytea,
    p_user_id uuid, p_expected_revision bigint, p_patch_profile boolean,
    p_profile_schema_id uuid, p_profile_document bytea, p_profile_sha256 bytea,
    p_replace_organizations boolean, p_organization_ids uuid[], p_membership_ids uuid[],
    p_replace_roles boolean, p_role_ids uuid[], p_assignment_ids uuid[],
    p_actor_id uuid, p_permission_id uuid, p_expected_token_version bigint,
    p_expected_authorization_version bigint
) RETURNS TABLE (result_id uuid, result_revision bigint,
    authorization_clock bigint, replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_status text; v_values jsonb; v_revision bigint; v_clock bigint;
    v_relations_changed boolean := false;
BEGIN
    IF p_command_key NOT IN ('iam.identity.user.create', 'iam.identity.user.update') OR
       p_user_id IS NULL OR p_expected_revision < 1 OR
       NOT (p_patch_profile OR p_replace_organizations OR p_replace_roles) OR
       ((p_replace_organizations OR p_replace_roles) AND p_user_id = p_actor_id) OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       (p_patch_profile AND (p_profile_schema_id IS NULL OR pg_catalog.octet_length(
          p_profile_sha256) IS DISTINCT FROM 32)) OR (NOT p_patch_profile AND
          (p_profile_schema_id IS NOT NULL OR p_profile_document IS NOT NULL OR
           p_profile_sha256 IS NOT NULL)) OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE =
          'IAM identity user graph authority is denied'; END IF;
    RETURN QUERY SELECT * FROM iam_identity_replay_v1(
        p_operation_id, p_command_key, p_request_sha256, p_user_id);
    IF FOUND THEN RETURN; END IF;
    IF (p_replace_organizations AND (COALESCE(pg_catalog.cardinality(
          p_organization_ids), -1) NOT BETWEEN 0 AND 100 OR p_membership_ids IS NULL OR
          pg_catalog.cardinality(p_organization_ids) <> pg_catalog.cardinality(
          p_membership_ids))) OR (NOT p_replace_organizations AND
          (COALESCE(pg_catalog.cardinality(p_organization_ids), 0) <> 0 OR COALESCE(
           pg_catalog.cardinality(p_membership_ids), 0) <> 0)) OR
       (p_replace_roles AND (COALESCE(pg_catalog.cardinality(p_role_ids), -1)
          NOT BETWEEN 0 AND 100 OR p_assignment_ids IS NULL OR pg_catalog.cardinality(
          p_role_ids) <> pg_catalog.cardinality(p_assignment_ids))) OR
       (NOT p_replace_roles AND (COALESCE(pg_catalog.cardinality(p_role_ids), 0) <> 0 OR
          COALESCE(pg_catalog.cardinality(p_assignment_ids), 0) <> 0)) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE =
            'IAM identity user graph budget is invalid'; END IF;
    IF (p_replace_organizations AND (
          p_organization_ids IS DISTINCT FROM ARRAY(
             SELECT value FROM pg_catalog.unnest(p_organization_ids) AS value ORDER BY value
          ) OR pg_catalog.cardinality(p_organization_ids) <> (SELECT pg_catalog.count(
             DISTINCT value) FROM pg_catalog.unnest(p_organization_ids) AS value) OR
          pg_catalog.cardinality(p_membership_ids) <> (SELECT pg_catalog.count(
             DISTINCT value) FROM pg_catalog.unnest(p_membership_ids) AS value) OR
          EXISTS (SELECT 1 FROM pg_catalog.generate_subscripts(
             p_organization_ids, 1) AS pair(position) WHERE p_organization_ids[pair.position] IS NULL OR
             p_membership_ids[pair.position] IS NULL))) OR
       (p_replace_roles AND (
          p_role_ids IS DISTINCT FROM ARRAY(
             SELECT value FROM pg_catalog.unnest(p_role_ids) AS value ORDER BY value
          ) OR pg_catalog.cardinality(p_role_ids) <> (SELECT pg_catalog.count(
             DISTINCT value) FROM pg_catalog.unnest(p_role_ids) AS value) OR
          pg_catalog.cardinality(p_assignment_ids) <> (SELECT pg_catalog.count(
             DISTINCT value) FROM pg_catalog.unnest(p_assignment_ids) AS value) OR
          EXISTS (SELECT 1 FROM pg_catalog.generate_subscripts(
             p_role_ids, 1) AS pair(position) WHERE p_role_ids[pair.position] IS NULL OR
             p_assignment_ids[pair.position] IS NULL))) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE =
            'IAM identity user graph set is not canonical'; END IF;
    SELECT principal.status, principal.revision INTO v_status, v_revision
    FROM iam_users AS principal WHERE principal.id = p_user_id FOR UPDATE;
    IF NOT FOUND OR v_revision <> p_expected_revision OR v_status = 'tombstoned' OR
       ((p_replace_organizations OR p_replace_roles) AND v_status <> 'active') THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE =
            'IAM identity user graph compare-and-swap failed'; END IF;
    IF p_patch_profile THEN
        v_values := iam_identity_profile_assert_v2(
            'user', p_profile_schema_id, p_profile_document, p_profile_sha256);
    END IF;
    IF p_replace_organizations AND (
       (SELECT pg_catalog.count(*) FROM iam_organizations WHERE id = ANY(
          p_organization_ids) AND status = 'active') <> pg_catalog.cardinality(
          p_organization_ids) OR
       EXISTS (SELECT 1 FROM iam_organization_memberships AS membership
        WHERE membership.user_id = p_user_id AND membership.membership_kind = 'owner'
          AND ((membership.state = 'active' AND NOT membership.organization_id = ANY(
               p_organization_ids)) OR (membership.state <> 'active' AND
               membership.organization_id = ANY(p_organization_ids)))) OR
       EXISTS (SELECT 1 FROM iam_organization_memberships AS membership
        WHERE membership.user_id = p_user_id AND membership.state = 'tombstoned'
          AND membership.organization_id = ANY(p_organization_ids))) THEN
        RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE =
            'IAM identity organization set is invalid'; END IF;
    IF p_replace_roles THEN
        IF (SELECT pg_catalog.count(*) FROM iam_roles WHERE id = ANY(
            p_role_ids) AND status = 'active') <> pg_catalog.cardinality(p_role_ids) OR
           EXISTS (SELECT 1 FROM iam_user_role_assignments AS assignment
            WHERE assignment.user_id = p_user_id AND assignment.state = 'revoked'
              AND assignment.role_id = ANY(p_role_ids)) THEN
            RAISE EXCEPTION USING ERRCODE = '23514', MESSAGE =
                'IAM identity role set is invalid'; END IF;
        IF (SELECT pg_catalog.count(*) FROM iam_user_role_assignments AS assignment
            JOIN iam_roles AS role ON role.id = assignment.role_id
            WHERE assignment.user_id = p_user_id AND assignment.state = 'active'
              AND role.is_protected) <> (SELECT pg_catalog.count(*) FROM iam_roles
            WHERE id = ANY(p_role_ids) AND is_protected) OR EXISTS (
            SELECT 1 FROM iam_user_role_assignments AS assignment
            JOIN iam_roles AS role ON role.id = assignment.role_id
            WHERE assignment.user_id = p_user_id AND assignment.state = 'active'
              AND role.is_protected AND NOT assignment.role_id = ANY(p_role_ids)) THEN
            RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE =
                'IAM protected role set is immutable'; END IF;
    END IF;
    IF p_replace_roles AND EXISTS (
       SELECT 1 FROM iam_roles AS role WHERE role.id = ANY(p_role_ids)
         AND role.scope_kind = 'organization' AND
         ((p_replace_organizations AND NOT role.scope_id = ANY(p_organization_ids)) OR
          (NOT p_replace_organizations AND NOT EXISTS (
             SELECT 1 FROM iam_organization_memberships AS membership
             WHERE membership.user_id = p_user_id
               AND membership.organization_id = role.scope_id AND membership.state = 'active')))
    ) OR p_replace_organizations AND EXISTS (
       SELECT 1 FROM iam_user_role_assignments AS assignment
       JOIN iam_roles AS role ON role.id = assignment.role_id
       WHERE assignment.user_id = p_user_id AND assignment.state = 'active'
         AND role.scope_kind = 'organization' AND NOT role.scope_id = ANY(p_organization_ids)
         AND (NOT p_replace_roles OR role.is_protected OR assignment.role_id = ANY(p_role_ids))
    ) THEN RAISE EXCEPTION USING ERRCODE = '23514',
        MESSAGE = 'IAM organization role membership is invalid'; END IF;
    v_relations_changed := (p_replace_organizations AND ((SELECT pg_catalog.count(*)
        FROM iam_organization_memberships WHERE user_id = p_user_id AND state = 'active') <>
        pg_catalog.cardinality(p_organization_ids) OR EXISTS (SELECT 1
        FROM iam_organization_memberships WHERE user_id = p_user_id AND state = 'active'
          AND NOT organization_id = ANY(p_organization_ids)))) OR (p_replace_roles AND
        ((SELECT pg_catalog.count(*) FROM iam_user_role_assignments
          WHERE user_id = p_user_id AND state = 'active') <> pg_catalog.cardinality(p_role_ids)
         OR EXISTS (SELECT 1 FROM iam_user_role_assignments WHERE user_id = p_user_id
          AND state = 'active' AND NOT role_id = ANY(p_role_ids))));
    IF p_replace_organizations THEN
        INSERT INTO iam_organization_memberships (
            id, organization_id, user_id, membership_kind, state,
            revision, created_at, updated_at
        ) SELECT p_membership_ids[desired.position], p_organization_ids[desired.position],
          p_user_id, 'member', 'active', 1, v_now, v_now FROM
          pg_catalog.generate_subscripts(p_organization_ids, 1) AS desired(position)
        ON CONFLICT (organization_id, user_id) DO UPDATE SET
            state = 'active', revision = iam_organization_memberships.revision + 1,
            tombstoned_at = NULL, updated_at = v_now
        WHERE iam_organization_memberships.membership_kind = 'member'
          AND iam_organization_memberships.state = 'suspended';
    END IF;
    IF p_replace_roles THEN
        UPDATE iam_user_role_assignments AS assignment SET state = 'suspended',
            revision = assignment.revision + 1, updated_at = v_now
        FROM iam_roles AS role WHERE role.id = assignment.role_id
          AND assignment.user_id = p_user_id AND assignment.state = 'active'
          AND NOT role.is_protected AND NOT assignment.role_id = ANY(p_role_ids);
        INSERT INTO iam_user_role_assignments (
            id, user_id, role_id, role_scope_kind, role_scope_id,
            organization_id, membership_id, state, revision, created_at, updated_at
        ) SELECT p_assignment_ids[desired.position], p_user_id, role.id,
            role.scope_kind, role.scope_id,
            CASE WHEN role.scope_kind = 'organization' THEN role.scope_id END,
            CASE WHEN role.scope_kind = 'organization' THEN membership.id END,
            'active', 1, v_now, v_now FROM pg_catalog.generate_subscripts(
            p_role_ids, 1) AS desired(position) JOIN iam_roles AS role
          ON role.id = p_role_ids[desired.position] AND NOT role.is_protected
          LEFT JOIN iam_organization_memberships AS membership ON
            membership.user_id = p_user_id AND membership.organization_id = role.scope_id
            AND membership.state = 'active'
        ON CONFLICT (user_id, role_id) DO UPDATE SET state = 'active',
            revision = iam_user_role_assignments.revision + 1,
            revoked_at = NULL, updated_at = v_now
        WHERE iam_user_role_assignments.state = 'suspended';
    END IF;
    IF p_replace_organizations THEN
        UPDATE iam_organization_memberships AS membership SET state = 'suspended',
            revision = membership.revision + 1, updated_at = v_now
        WHERE membership.user_id = p_user_id AND membership.membership_kind = 'member'
          AND membership.state = 'active'
          AND NOT membership.organization_id = ANY(p_organization_ids);
    END IF;
    UPDATE iam_users SET
        profile_schema_version_id = CASE WHEN p_patch_profile
            THEN p_profile_schema_id ELSE profile_schema_version_id END,
        profile_values = CASE WHEN p_patch_profile THEN v_values ELSE profile_values END,
        profile_sha256 = CASE WHEN p_patch_profile THEN p_profile_sha256 ELSE profile_sha256 END,
        authorization_version = authorization_version +
            CASE WHEN v_relations_changed THEN 1 ELSE 0 END,
        revision = revision + 1, updated_at = v_now
    WHERE id = p_user_id AND revision = p_expected_revision RETURNING revision INTO v_revision;
    v_clock := iam_identity_record_v1(p_operation_id, p_command_key, p_request_sha256,
        p_user_id, v_revision, p_request_sha256);
    RETURN QUERY SELECT p_user_id, v_revision, v_clock, false;
END; $function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_identity_user_relations_replace_v1(name,uuid,text,bytea,uuid,
    bigint,boolean,uuid,bytea,bytea,boolean,uuid[],uuid[],boolean,uuid[],uuid[],uuid,uuid,bigint,bigint) FROM PUBLIC;
