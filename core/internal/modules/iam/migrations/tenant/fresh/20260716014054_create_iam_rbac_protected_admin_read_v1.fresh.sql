-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_protected_admin_get_v1(
    p_scope_schema name, p_role_id uuid, p_actor_id uuid, p_required_permission_id uuid
)
RETURNS TABLE (role_id uuid, tenant_all_permission_id uuid,
    role_permission_binding_id uuid, minimum_active_assignments integer,
    authorization_epoch bigint, revision bigint)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT iam_rbac_read_assert_v1(p_scope_schema, p_actor_id, p_required_permission_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'IAM RBAC query denied';
    END IF;
    RETURN QUERY SELECT policy.role_id, policy.tenant_all_permission_id,
        policy.role_permission_binding_id, policy.minimum_active_assignments,
        policy.authorization_epoch, policy.revision
    FROM iam_protected_admin_policies AS policy
    WHERE policy.singleton AND policy.policy_state = 'active' AND policy.role_id = p_role_id;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_rbac_protected_admin_get_v1(name,uuid,uuid,uuid) FROM PUBLIC;
