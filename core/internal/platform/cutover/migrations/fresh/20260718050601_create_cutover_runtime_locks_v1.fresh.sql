-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.cutover_runtime_lock_group_v1(
    p_scope_kind text,
    p_scope_id uuid,
    p_group_id text
)
RETURNS TABLE (state text, writer text, generation bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Cutover runtime group lock denied';
    END IF;
    RETURN QUERY SELECT group_row.state, group_row.writer, group_row.generation
    FROM platform.cutover_groups AS group_row
    WHERE group_row.scope_kind = p_scope_kind
      AND group_row.scope_id = p_scope_id
      AND group_row.group_id = p_group_id
    ORDER BY group_row.scope_kind COLLATE "C", group_row.scope_id,
        group_row.group_id COLLATE "C"
    FOR SHARE;
END;
$function$;
CREATE FUNCTION platform.cutover_runtime_lock_group_members_v1(
    p_scope_kind text,
    p_scope_id uuid,
    p_group_id text
)
RETURNS TABLE (domain text)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Cutover runtime group member lock denied';
    END IF;
    RETURN QUERY SELECT member.domain
    FROM platform.cutover_group_members AS member
    WHERE member.scope_kind = p_scope_kind
      AND member.scope_id = p_scope_id
      AND member.group_id = p_group_id
    ORDER BY member.domain COLLATE "C"
    FOR SHARE;
END;
$function$;
CREATE FUNCTION platform.cutover_runtime_lock_domain_v1(
    p_scope_kind text,
    p_scope_id uuid,
    p_domain text
)
RETURNS TABLE (state text, writer text, generation bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Cutover runtime domain lock denied';
    END IF;
    RETURN QUERY SELECT domain_row.state, domain_row.writer, domain_row.generation
    FROM platform.domain_cutovers AS domain_row
    WHERE domain_row.scope_kind = p_scope_kind
      AND domain_row.scope_id = p_scope_id
      AND domain_row.domain = p_domain
    ORDER BY domain_row.scope_kind COLLATE "C", domain_row.scope_id,
        domain_row.domain COLLATE "C"
    FOR SHARE;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.cutover_runtime_lock_group_v1(text, uuid, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.cutover_runtime_lock_group_v1(text, uuid, text) TO forgecrud_runtime;
REVOKE ALL ON FUNCTION
    platform.cutover_runtime_lock_group_members_v1(text, uuid, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.cutover_runtime_lock_group_members_v1(text, uuid, text) TO forgecrud_runtime;
REVOKE ALL ON FUNCTION
    platform.cutover_runtime_lock_domain_v1(text, uuid, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.cutover_runtime_lock_domain_v1(text, uuid, text) TO forgecrud_runtime;
