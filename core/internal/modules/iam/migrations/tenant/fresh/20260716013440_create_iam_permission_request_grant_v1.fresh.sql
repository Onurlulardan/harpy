-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_grant_v1(
    p_request_id uuid,
    p_now timestamptz
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_target_kind text;
    v_target_user_id uuid;
    v_target_role_id uuid;
    v_target_organization_id uuid;
    v_permission_id uuid;
    v_planned_binding_id uuid;
    v_role_is_protected boolean;
BEGIN
    SELECT request.target_kind, request.target_user_id,
           request.target_role_id, request.target_organization_id,
           request.permission_id, request.planned_binding_id
    INTO v_target_kind, v_target_user_id, v_target_role_id,
         v_target_organization_id, v_permission_id, v_planned_binding_id
    FROM iam_permission_requests AS request
    JOIN iam_permissions AS permission ON permission.id = request.permission_id
    WHERE request.id OPERATOR(pg_catalog.=) p_request_id
      AND request.state OPERATOR(pg_catalog.=) 'IN_REVIEW'
      AND permission.match_kind OPERATOR(pg_catalog.=) 'exact'
      AND permission.state OPERATOR(pg_catalog.=) 'active'
    FOR UPDATE OF request;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM permission request grant evidence is stale';
    END IF;
    IF v_target_kind OPERATOR(pg_catalog.=) 'USER' THEN
        IF NOT EXISTS (
            SELECT 1 FROM iam_users AS principal
            WHERE principal.id = v_target_user_id
              AND principal.status = 'active'
        ) OR EXISTS (
            SELECT 1 FROM iam_user_permission_bindings AS binding
            WHERE binding.user_id = v_target_user_id
              AND binding.permission_id = v_permission_id
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM permission request user grant is stale';
        END IF;
        INSERT INTO iam_user_permission_bindings (
            id, user_id, permission_id, permission_match_kind, state,
            revision, created_at, updated_at, activated_at, revoked_at
        ) VALUES (
            v_planned_binding_id, v_target_user_id,
            v_permission_id, 'exact', 'active', 1,
            p_now, p_now, p_now, NULL
        );
    ELSIF v_target_kind OPERATOR(pg_catalog.=) 'ROLE' THEN
        SELECT role.is_protected INTO v_role_is_protected
        FROM iam_roles AS role
        WHERE role.id = v_target_role_id AND role.status = 'active'
        FOR UPDATE;
        IF NOT FOUND OR EXISTS (
            SELECT 1 FROM iam_role_permission_bindings AS binding
            WHERE binding.role_id = v_target_role_id
              AND binding.permission_id = v_permission_id
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM permission request role grant is stale';
        END IF;
        INSERT INTO iam_role_permission_bindings (
            id, role_id, role_is_protected, permission_id,
            permission_match_kind, state, revision, created_at,
            updated_at, activated_at, revoked_at
        ) VALUES (
            v_planned_binding_id, v_target_role_id,
            v_role_is_protected, v_permission_id, 'exact', 'active',
            1, p_now, p_now, p_now, NULL
        );
    ELSIF v_target_kind OPERATOR(pg_catalog.=) 'ORGANIZATION' THEN
        IF NOT EXISTS (
            SELECT 1 FROM iam_organizations AS organization
            WHERE organization.id = v_target_organization_id
              AND organization.status = 'active'
        ) OR EXISTS (
            SELECT 1 FROM iam_organization_permission_bindings AS binding
            WHERE binding.organization_id = v_target_organization_id
              AND binding.permission_id = v_permission_id
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM permission request organization grant is stale';
        END IF;
        INSERT INTO iam_organization_permission_bindings (
            id, organization_id, permission_id, permission_match_kind,
            state, revision, created_at, updated_at, activated_at, revoked_at
        ) VALUES (
            v_planned_binding_id, v_target_organization_id,
            v_permission_id, 'exact', 'active', 1,
            p_now, p_now, p_now, NULL
        );
    ELSE
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM permission request target kind is invalid';
    END IF;
    RETURN v_planned_binding_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_grant_v1(uuid,timestamptz)
    FROM PUBLIC;
