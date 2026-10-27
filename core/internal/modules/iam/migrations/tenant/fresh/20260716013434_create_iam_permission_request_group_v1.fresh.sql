-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_group_create_v1(
    p_scope_schema name,
    p_group_id uuid,
    p_requester_user_id uuid,
    p_request_mode text,
    p_creator_authority_permission_id uuid,
    p_reviewer_authority_permission_id uuid,
    p_expected_authorization_version bigint,
    p_step_up_purpose_sha256 bytea,
    p_expires_at timestamptz,
    p_requests jsonb
)
RETURNS TABLE (group_id uuid, group_state text, group_revision bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_user_state text;
    v_authorization_version bigint;
    v_request_count integer;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission request creation scope is invalid';
    END IF;
    IF p_group_id IS NULL OR p_requester_user_id IS NULL OR
       p_request_mode NOT IN ('SELF_SERVICE', 'PRIVILEGED') OR
       p_creator_authority_permission_id IS NULL OR
       p_reviewer_authority_permission_id IS NULL OR
       p_expected_authorization_version < 1 OR
       pg_catalog.octet_length(p_step_up_purpose_sha256) IS DISTINCT FROM 32 OR
       p_expires_at <= v_now OR pg_catalog.jsonb_typeof(p_requests) <> 'array' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM permission request creation input is invalid';
    END IF;
    v_request_count := pg_catalog.jsonb_array_length(p_requests);
    IF v_request_count NOT BETWEEN 1 AND 100 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM permission request group size is invalid';
    END IF;
    SELECT principal.status, principal.authorization_version
    INTO v_user_state, v_authorization_version
    FROM iam_users AS principal
    WHERE principal.id OPERATOR(pg_catalog.=) p_requester_user_id
    FOR UPDATE;
    IF NOT FOUND OR v_user_state IS DISTINCT FROM 'active' OR
       v_authorization_version IS DISTINCT FROM p_expected_authorization_version THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM permission request creator evidence is stale';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM iam_permissions AS permission
        WHERE permission.id IN (
            p_creator_authority_permission_id,
            p_reviewer_authority_permission_id
        ) AND permission.match_kind = 'exact' AND permission.is_protected
          AND permission.state = 'active'
        HAVING pg_catalog.count(*) = 2
    ) OR (p_request_mode = 'PRIVILEGED' AND NOT iam_user_has_permission_v1(
        p_requester_user_id, p_creator_authority_permission_id
    )) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission request creator authority is denied';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(p_requests) AS document(value)
        WHERE pg_catalog.jsonb_typeof(document.value) <> 'object' OR
          (SELECT pg_catalog.count(*)
           FROM pg_catalog.jsonb_object_keys(document.value)) <> 6
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM permission request item shape is invalid';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_to_recordset(p_requests) AS item(
            id uuid, ordinal smallint, target_kind text, target_id uuid,
            permission_id uuid, planned_binding_id uuid
        )
        LEFT JOIN iam_permissions AS permission ON permission.id = item.permission_id
        LEFT JOIN iam_users AS target_user
            ON item.target_kind = 'USER' AND target_user.id = item.target_id
        LEFT JOIN iam_roles AS target_role
            ON item.target_kind = 'ROLE' AND target_role.id = item.target_id
        LEFT JOIN iam_organizations AS target_organization
            ON item.target_kind = 'ORGANIZATION' AND
               target_organization.id = item.target_id
        WHERE item.id IS NULL OR item.target_id IS NULL OR
          item.permission_id IS NULL OR item.planned_binding_id IS NULL OR
          item.ordinal NOT BETWEEN 1 AND v_request_count OR
          item.target_kind NOT IN ('USER', 'ROLE', 'ORGANIZATION') OR
          permission.match_kind IS DISTINCT FROM 'exact' OR
          permission.state IS DISTINCT FROM 'active' OR
          (item.target_kind = 'USER' AND target_user.status IS DISTINCT FROM 'active') OR
          (item.target_kind = 'ROLE' AND target_role.status IS DISTINCT FROM 'active') OR
          (item.target_kind = 'ORGANIZATION' AND
              target_organization.status IS DISTINCT FROM 'active') OR
          (p_request_mode = 'SELF_SERVICE' AND (
              item.target_kind <> 'USER' OR
              item.target_id IS DISTINCT FROM p_requester_user_id OR
              permission.is_protected
          ))
    ) OR (SELECT pg_catalog.count(DISTINCT item.ordinal)
          FROM pg_catalog.jsonb_to_recordset(p_requests) AS item(ordinal smallint))
          <> v_request_count THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM permission request item is invalid';
    END IF;
    INSERT INTO iam_permission_request_groups (
        id, requester_user_id, request_mode, creator_authority_permission_id,
        reviewer_authority_permission_id, step_up_purpose_sha256, state,
        request_count, reviewer_user_id, review_task_id, revision, expires_at,
        terminal_at, created_at, updated_at
    ) VALUES (
        p_group_id, p_requester_user_id, p_request_mode,
        p_creator_authority_permission_id, p_reviewer_authority_permission_id,
        p_step_up_purpose_sha256, 'PENDING', v_request_count,
        NULL, NULL, 1, p_expires_at, NULL, v_now, v_now
    );
    INSERT INTO iam_permission_requests (
        id, group_id, ordinal, target_kind, target_user_id, target_role_id,
        target_organization_id, permission_id, permission_match_kind,
        permission_is_protected, planned_binding_id, granted_binding_id,
        state, revision, terminal_at, created_at, updated_at
    )
    SELECT item.id, p_group_id, item.ordinal, item.target_kind,
        CASE WHEN item.target_kind = 'USER' THEN item.target_id END,
        CASE WHEN item.target_kind = 'ROLE' THEN item.target_id END,
        CASE WHEN item.target_kind = 'ORGANIZATION' THEN item.target_id END,
        permission.id, permission.match_kind, permission.is_protected,
        item.planned_binding_id, NULL, 'PENDING', 1, NULL, v_now, v_now
    FROM pg_catalog.jsonb_to_recordset(p_requests) AS item(
        id uuid, ordinal smallint, target_kind text, target_id uuid,
        permission_id uuid, planned_binding_id uuid
    )
    JOIN iam_permissions AS permission ON permission.id = item.permission_id;
    RETURN QUERY SELECT p_group_id, 'PENDING'::text, 1::bigint;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_group_create_v1(
    name,uuid,uuid,text,uuid,uuid,bigint,bytea,timestamptz,jsonb
) FROM PUBLIC;
