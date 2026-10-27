-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_implication_get_v1(
    p_scope_schema name, p_id uuid, p_actor_id uuid, p_required_permission_id uuid
)
RETURNS TABLE (id uuid, source_permission_id uuid, implied_permission_id uuid,
    retirement_contract_key text, state text, hit_count bigint, revision bigint)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT iam_rbac_read_assert_v1(p_scope_schema, p_actor_id, p_required_permission_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'IAM RBAC query denied';
    END IF;
    RETURN QUERY SELECT edge.id, edge.source_permission_id, edge.implied_permission_id,
        edge.retirement_contract_key, edge.state, edge.hit_count, edge.revision
    FROM iam_permission_implications AS edge WHERE edge.id = p_id;
END;
$function$;

CREATE FUNCTION iam_rbac_implication_list_v1(
    p_scope_schema name, p_state text, p_limit integer,
    p_actor_id uuid, p_required_permission_id uuid
)
RETURNS TABLE (id uuid, source_permission_id uuid, implied_permission_id uuid,
    retirement_contract_key text, state text, hit_count bigint, revision bigint)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_limit < 1 OR p_limit > 200 OR p_state NOT IN ('active','retired','all') THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'IAM RBAC implication query is invalid';
    END IF;
    IF NOT iam_rbac_read_assert_v1(p_scope_schema, p_actor_id, p_required_permission_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'IAM RBAC query denied';
    END IF;
    RETURN QUERY SELECT edge.id, edge.source_permission_id, edge.implied_permission_id,
        edge.retirement_contract_key, edge.state, edge.hit_count, edge.revision
    FROM iam_permission_implications AS edge
    WHERE p_state = 'all' OR edge.state = p_state
    ORDER BY edge.retirement_contract_key, edge.id LIMIT p_limit;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_rbac_implication_get_v1(name,uuid,uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_rbac_implication_list_v1(name,text,integer,uuid,uuid) FROM PUBLIC;
