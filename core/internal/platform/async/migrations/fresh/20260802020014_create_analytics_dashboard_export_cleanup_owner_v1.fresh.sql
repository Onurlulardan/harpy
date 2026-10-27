-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
    p_scope_kind text,p_scope_id uuid,p_authority_group_id text,
    p_authority_domain text
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       p_scope_kind IS DISTINCT FROM 'tenant' OR p_scope_id IS NULL OR
       p_authority_group_id !~ '^[a-z][a-z0-9_]{0,62}$' OR
       p_authority_domain IS DISTINCT FROM 'analytics' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export cleanup owner input is invalid';
    END IF;
    PERFORM 1 FROM platform.cutover_groups AS group_row
    JOIN platform.domain_cutovers AS domain_row
      ON domain_row.scope_kind=group_row.scope_kind
     AND domain_row.scope_id=group_row.scope_id
    JOIN platform.cutover_group_members AS member
      ON member.scope_kind=group_row.scope_kind
     AND member.scope_id=group_row.scope_id
     AND member.group_id=group_row.group_id
     AND member.domain=domain_row.domain
    WHERE group_row.scope_kind=p_scope_kind
      AND group_row.scope_id=p_scope_id
      AND group_row.group_id=p_authority_group_id
      AND group_row.state='v2' AND group_row.writer='v2'
      AND domain_row.domain=p_authority_domain
      AND domain_row.state='v2' AND domain_row.writer='v2';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export cleanup owner is not current';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
        text,uuid,text,text)
FROM PUBLIC;
