-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_access_v1(
    p_actor_id uuid, p_dashboard_id uuid, p_version_id uuid
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM analytics_dashboards AS root
        JOIN analytics_dashboard_versions AS version_row
          ON version_row.dashboard_id=root.id
        WHERE root.id=p_dashboard_id
          AND version_row.id=p_version_id
          AND (
              version_row.visibility='tenant'
              OR root.created_by=p_actor_id
              OR EXISTS (
                  SELECT 1 FROM analytics_dashboard_acls AS acl
                  WHERE acl.dashboard_id=root.id
                    AND acl.principal_kind='user'
                    AND acl.principal_id=p_actor_id
                    AND acl.access_level IN ('read','edit','publish')
              )
          )
    );
END;
$function$;

CREATE FUNCTION analytics_dashboard_export_snapshot_v1(
    p_scope_schema name, p_actor_id uuid, p_dashboard_id uuid,
    p_export_id uuid, p_observed_at timestamptz
)
RETURNS TABLE (
    export_id uuid, dashboard_id uuid, dashboard_version_id uuid, job_id uuid,
    effective_state text, stored_state text, format text, filename text,
    content_type text, failure_code text, row_count bigint,
    plaintext_bytes bigint, cancellation_requested boolean,
    created_at timestamptz, updated_at timestamptz, expires_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_actor_id IS NULL OR
       p_actor_id='00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.uuid_extract_version(p_dashboard_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='analytics dashboard export read input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    RETURN QUERY
    SELECT item.export_id,item.dashboard_id,item.dashboard_version_id,item.job_id,
           CASE WHEN item.state='ready' AND item.expires_at<=p_observed_at
                THEN 'expired' ELSE item.state END,
           item.state,item.format,COALESCE(item.filename,''),
           COALESCE(item.content_type,''),COALESCE(item.failure_code,''),
           COALESCE(item.row_count,0),COALESCE(item.plaintext_bytes,0),
           item.cancellation_requested_at IS NOT NULL,
           item.created_at,item.updated_at,item.expires_at
    FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id
      AND item.dashboard_id=p_dashboard_id
      AND analytics_dashboard_export_access_v1(
          p_actor_id,item.dashboard_id,item.dashboard_version_id);
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics dashboard export was not found';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_access_v1(uuid,uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_snapshot_v1(
    name,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
