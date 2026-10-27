-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_resource_get_v1(
    p_scope_schema name, p_id uuid, p_actor_id uuid, p_required_permission_id uuid
)
RETURNS TABLE (
    id uuid, resource_key text, owner_domain text, resource_kind text,
    display_name text, description text, is_protected boolean, state text, revision bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT iam_rbac_read_assert_v1(p_scope_schema, p_actor_id, p_required_permission_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'IAM RBAC query denied';
    END IF;
    RETURN QUERY SELECT resource.id, resource.resource_key, resource.owner_domain,
        resource.resource_kind, resource.display_name, resource.description,
        resource.is_protected, resource.state, resource.revision
    FROM iam_resources AS resource WHERE resource.id = p_id;
END;
$function$;

CREATE FUNCTION iam_rbac_resource_list_v1(
    p_scope_schema name, p_state text, p_limit integer,
    p_actor_id uuid, p_required_permission_id uuid
)
RETURNS TABLE (
    id uuid, resource_key text, owner_domain text, resource_kind text,
    display_name text, description text, is_protected boolean, state text, revision bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_limit < 1 OR p_limit > 200 OR p_state NOT IN ('active','retired','all') THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'IAM RBAC resource query is invalid';
    END IF;
    IF NOT iam_rbac_read_assert_v1(p_scope_schema, p_actor_id, p_required_permission_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'IAM RBAC query denied';
    END IF;
    RETURN QUERY SELECT resource.id, resource.resource_key, resource.owner_domain,
        resource.resource_kind, resource.display_name, resource.description,
        resource.is_protected, resource.state, resource.revision
    FROM iam_resources AS resource
    WHERE p_state = 'all' OR resource.state = p_state
    ORDER BY resource.resource_key, resource.id LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_rbac_resource_get_v1(name,uuid,uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_rbac_resource_list_v1(name,text,integer,uuid,uuid) FROM PUBLIC;
