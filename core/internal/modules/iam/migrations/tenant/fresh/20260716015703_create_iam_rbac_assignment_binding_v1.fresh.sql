-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_assignment_binding_grant_v1(
    p_group_id uuid, p_target_kind text, p_target_id uuid,
    p_permission_id uuid, p_suggested_binding_id uuid
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_binding_id uuid; v_state text; v_protected boolean;
    v_owned boolean; v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF p_suggested_binding_id IS NULL OR NOT EXISTS (
        SELECT 1 FROM iam_permissions
        WHERE id = p_permission_id AND match_kind = 'exact' AND state = 'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM RBAC assignment binding input is invalid';
    END IF;
    CASE p_target_kind
    WHEN 'USER' THEN
        SELECT id, state INTO v_binding_id, v_state
        FROM iam_user_permission_bindings
        WHERE user_id = p_target_id AND permission_id = p_permission_id
        FOR UPDATE;
    WHEN 'ROLE' THEN
        SELECT id, state INTO v_binding_id, v_state
        FROM iam_role_permission_bindings
        WHERE role_id = p_target_id AND permission_id = p_permission_id
        FOR UPDATE;
    WHEN 'ORGANIZATION' THEN
        SELECT id, state INTO v_binding_id, v_state
        FROM iam_organization_permission_bindings
        WHERE organization_id = p_target_id AND permission_id = p_permission_id
        FOR UPDATE;
    ELSE
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM RBAC assignment target is invalid';
    END CASE;
    IF v_binding_id IS NOT NULL THEN
        SELECT true INTO v_owned FROM iam_rbac_assignment_members
        WHERE group_id = p_group_id AND binding_id = v_binding_id;
        IF v_owned IS DISTINCT FROM true THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM RBAC binding is owned outside the compatibility group';
        END IF;
        IF v_state = 'revoked' THEN
            CASE p_target_kind
            WHEN 'USER' THEN
                UPDATE iam_user_permission_bindings SET state = 'active',
                    revision = revision + 1, updated_at = v_now,
                    activated_at = v_now, revoked_at = NULL
                WHERE id = v_binding_id;
            WHEN 'ROLE' THEN
                UPDATE iam_role_permission_bindings SET state = 'active',
                    revision = revision + 1, updated_at = v_now,
                    activated_at = v_now, revoked_at = NULL
                WHERE id = v_binding_id;
            WHEN 'ORGANIZATION' THEN
                UPDATE iam_organization_permission_bindings SET state = 'active',
                    revision = revision + 1, updated_at = v_now,
                    activated_at = v_now, revoked_at = NULL
                WHERE id = v_binding_id;
            END CASE;
        END IF;
        RETURN v_binding_id;
    END IF;
    v_binding_id := p_suggested_binding_id;
    CASE p_target_kind
    WHEN 'USER' THEN
        INSERT INTO iam_user_permission_bindings (
            id, user_id, permission_id, permission_match_kind, state, revision,
            created_at, updated_at, activated_at, revoked_at
        ) VALUES (
            v_binding_id, p_target_id, p_permission_id, 'exact', 'active', 1,
            v_now, v_now, v_now, NULL
        );
    WHEN 'ROLE' THEN
        SELECT is_protected INTO v_protected FROM iam_roles
        WHERE id = p_target_id AND status = 'active';
        IF v_protected IS NULL THEN
            RAISE EXCEPTION USING ERRCODE = '42501',
                MESSAGE = 'IAM RBAC assignment role is unavailable';
        END IF;
        INSERT INTO iam_role_permission_bindings (
            id, role_id, role_is_protected, permission_id,
            permission_match_kind, state, revision,
            created_at, updated_at, activated_at, revoked_at
        ) VALUES (
            v_binding_id, p_target_id, v_protected, p_permission_id,
            'exact', 'active', 1, v_now, v_now, v_now, NULL
        );
    WHEN 'ORGANIZATION' THEN
        INSERT INTO iam_organization_permission_bindings (
            id, organization_id, permission_id, permission_match_kind,
            state, revision, created_at, updated_at, activated_at, revoked_at
        ) VALUES (
            v_binding_id, p_target_id, p_permission_id, 'exact',
            'active', 1, v_now, v_now, v_now, NULL
        );
    END CASE;
    RETURN v_binding_id;
END;
$function$;
REVOKE ALL ON FUNCTION iam_rbac_assignment_binding_grant_v1(
    uuid,text,uuid,uuid,uuid
) FROM PUBLIC;

CREATE FUNCTION iam_rbac_assignment_binding_revoke_v1(
    p_target_kind text, p_target_id uuid,
    p_permission_id uuid, p_binding_id uuid
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE v_count bigint; v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    CASE p_target_kind
    WHEN 'USER' THEN
        UPDATE iam_user_permission_bindings SET state = 'revoked',
            revision = revision + 1, updated_at = v_now, revoked_at = v_now
        WHERE id = p_binding_id AND user_id = p_target_id
          AND permission_id = p_permission_id AND state = 'active';
    WHEN 'ROLE' THEN
        UPDATE iam_role_permission_bindings SET state = 'revoked',
            revision = revision + 1, updated_at = v_now, revoked_at = v_now
        WHERE id = p_binding_id AND role_id = p_target_id
          AND permission_id = p_permission_id AND state = 'active';
    WHEN 'ORGANIZATION' THEN
        UPDATE iam_organization_permission_bindings SET state = 'revoked',
            revision = revision + 1, updated_at = v_now, revoked_at = v_now
        WHERE id = p_binding_id AND organization_id = p_target_id
          AND permission_id = p_permission_id AND state = 'active';
    ELSE
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM RBAC assignment target is invalid';
    END CASE;
    GET DIAGNOSTICS v_count = ROW_COUNT;
    IF v_count IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM RBAC assignment binding changed';
    END IF;
    RETURN true;
END;
$function$;
REVOKE ALL ON FUNCTION iam_rbac_assignment_binding_revoke_v1(
    text,uuid,uuid,uuid
) FROM PUBLIC;

CREATE FUNCTION iam_rbac_assignment_bump_target_v1(
    p_target_kind text, p_target_id uuid, p_clock bigint
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    CASE p_target_kind
    WHEN 'USER' THEN
        RETURN iam_rbac_bump_user_v1(p_target_id, p_clock);
    WHEN 'ROLE' THEN
        RETURN iam_rbac_bump_role_users_v1(p_target_id, p_clock);
    WHEN 'ORGANIZATION' THEN
        RETURN iam_rbac_bump_organization_users_v1(p_target_id, p_clock);
    ELSE
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM RBAC assignment target is invalid';
    END CASE;
END;
$function$;
REVOKE ALL ON FUNCTION iam_rbac_assignment_bump_target_v1(
    text,uuid,bigint
) FROM PUBLIC;
-- +goose StatementEnd

