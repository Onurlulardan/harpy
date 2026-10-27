-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_organization_binding_get_v1(
    p_scope_schema name, p_id uuid, p_actor_id uuid, p_required_permission_id uuid
)
RETURNS TABLE (id uuid, subject_id uuid, permission_id uuid, state text, revision bigint)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT iam_rbac_read_assert_v1(p_scope_schema, p_actor_id, p_required_permission_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'IAM RBAC query denied';
    END IF;
    RETURN QUERY SELECT binding.id, binding.organization_id, binding.permission_id,
        binding.state, binding.revision
    FROM iam_organization_permission_bindings AS binding WHERE binding.id = p_id;
END;
$function$;

CREATE FUNCTION iam_rbac_organization_binding_list_v1(
    p_scope_schema name, p_organization_id uuid, p_state text, p_limit integer,
    p_actor_id uuid, p_required_permission_id uuid
)
RETURNS TABLE (id uuid, subject_id uuid, permission_id uuid, state text, revision bigint)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_limit < 1 OR p_limit > 200 OR p_state NOT IN ('active','revoked','all') THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'IAM RBAC binding query is invalid';
    END IF;
    IF NOT iam_rbac_read_assert_v1(p_scope_schema, p_actor_id, p_required_permission_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'IAM RBAC query denied';
    END IF;
    RETURN QUERY SELECT binding.id, binding.organization_id, binding.permission_id,
        binding.state, binding.revision FROM iam_organization_permission_bindings AS binding
    WHERE binding.organization_id = p_organization_id
      AND (p_state = 'all' OR binding.state = p_state)
    ORDER BY binding.permission_id, binding.id LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_rbac_organization_binding_get_v1(
    name,uuid,uuid,uuid
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_rbac_organization_binding_list_v1(
    name,uuid,text,integer,uuid,uuid
) FROM PUBLIC;
