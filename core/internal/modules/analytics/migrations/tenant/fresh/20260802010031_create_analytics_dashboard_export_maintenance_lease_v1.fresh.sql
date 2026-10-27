-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_maintenance_unknown_checkpoint_v1(
    p_scope_schema name,p_kind text,p_tenant_id uuid,p_tenant_schema name,
    p_export_id uuid,p_authority_group_id text,p_authority_group_generation bigint,
    p_cutover_generation bigint,p_job_id uuid,p_attempt_id uuid,p_attempt_number integer,
    p_lease_generation bigint,p_original_lease_owner_id uuid,p_effect_id uuid,
    p_request_sha256 bytea,p_unknown_at timestamptz,p_cleanup_generation bigint,
    p_cleanup_lease_owner_id uuid,p_delete_attempt_generation bigint,
    p_observed_at timestamptz
)
RETURNS timestamptz
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_tenant_id uuid; v_lease_until timestamptz; v_export analytics_dashboard_exports%ROWTYPE;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity WHERE singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_cleanup_generation<1 OR
       p_delete_attempt_generation NOT BETWEEN 0 AND 3 OR
       pg_catalog.uuid_extract_version(p_cleanup_lease_owner_id) IS DISTINCT FROM 7 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export maintenance checkpoint scope is invalid';
    END IF;
    v_lease_until:=
        platform.platform_async_analytics_dashboard_export_maintenance_unknown_checkpoint_v1(
            p_kind,p_tenant_id,p_tenant_schema,p_export_id,p_authority_group_id,
            p_authority_group_generation,p_cutover_generation,p_job_id,p_attempt_id,
            p_attempt_number,p_lease_generation,p_original_lease_owner_id,p_effect_id,
            p_request_sha256,p_unknown_at,p_cleanup_generation,
            p_cleanup_lease_owner_id,p_observed_at);
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id
      AND item.maintenance_unknown_job_id=p_job_id
      AND item.maintenance_cleanup_generation=p_cleanup_generation
      AND item.maintenance_cleanup_lease_owner=p_cleanup_lease_owner_id
      AND item.delete_attempt_generation=p_delete_attempt_generation
      AND item.maintenance_local_state_sha256 IS NOT NULL FOR UPDATE;
    IF v_export.export_id IS NULL OR v_lease_until<=p_observed_at OR
       (p_kind='analytics.dashboard.export.delete' AND
        v_export.delete_job_id IS DISTINCT FROM p_job_id) OR
       (p_kind='analytics.dashboard.export.expire' AND
        v_export.expiry_job_id IS DISTINCT FROM p_job_id) OR
       (v_export.state='deleting' AND v_export.quota_state='used' AND
        (v_export.delete_lease_owner IS DISTINCT FROM p_cleanup_lease_owner_id OR
         v_export.delete_lease_until IS DISTINCT FROM
            v_export.maintenance_cleanup_lease_until)) THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export maintenance checkpoint lease is stale';
    END IF;
    UPDATE analytics_dashboard_exports
    SET maintenance_cleanup_lease_until=v_lease_until,
        delete_lease_until=CASE WHEN delete_lease_owner=p_cleanup_lease_owner_id
            THEN v_lease_until ELSE delete_lease_until END,
        row_revision=row_revision+1,
        updated_at=GREATEST(p_observed_at,updated_at+interval '1 microsecond')
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    RETURN v_lease_until;
END;
$function$;

CREATE FUNCTION analytics_dashboard_export_maintenance_unknown_defer_v1(
    p_scope_schema name,p_kind text,p_tenant_id uuid,p_tenant_schema name,
    p_export_id uuid,p_authority_group_id text,p_authority_group_generation bigint,
    p_cutover_generation bigint,p_job_id uuid,p_attempt_id uuid,p_attempt_number integer,
    p_lease_generation bigint,p_original_lease_owner_id uuid,p_effect_id uuid,
    p_request_sha256 bytea,p_unknown_at timestamptz,p_cleanup_generation bigint,
    p_cleanup_lease_owner_id uuid,p_delete_attempt_generation bigint,
    p_observed_at timestamptz
)
RETURNS timestamptz
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_tenant_id uuid; v_retry_at timestamptz; v_export analytics_dashboard_exports%ROWTYPE;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity WHERE singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_cleanup_generation<1 OR
       p_delete_attempt_generation NOT BETWEEN 0 AND 3 OR
       pg_catalog.uuid_extract_version(p_cleanup_lease_owner_id) IS DISTINCT FROM 7 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export maintenance defer scope is invalid';
    END IF;
    v_retry_at:=
        platform.platform_async_analytics_dashboard_export_maintenance_unknown_defer_v1(
            p_kind,p_tenant_id,p_tenant_schema,p_export_id,p_authority_group_id,
            p_authority_group_generation,p_cutover_generation,p_job_id,p_attempt_id,
            p_attempt_number,p_lease_generation,p_original_lease_owner_id,p_effect_id,
            p_request_sha256,p_unknown_at,p_cleanup_generation,
            p_cleanup_lease_owner_id,p_observed_at);
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=p_tenant_id
      AND item.maintenance_unknown_job_id=p_job_id
      AND item.maintenance_cleanup_generation=p_cleanup_generation
      AND item.maintenance_cleanup_lease_owner=p_cleanup_lease_owner_id
      AND item.delete_attempt_generation=p_delete_attempt_generation
      AND item.maintenance_local_state_sha256 IS NOT NULL FOR UPDATE;
    IF v_export.export_id IS NULL OR
       (p_kind='analytics.dashboard.export.delete' AND
        v_export.delete_job_id IS DISTINCT FROM p_job_id) OR
       (p_kind='analytics.dashboard.export.expire' AND
        v_export.expiry_job_id IS DISTINCT FROM p_job_id) OR
       (v_export.state='deleting' AND v_export.quota_state='used' AND
        (v_export.delete_lease_owner IS DISTINCT FROM p_cleanup_lease_owner_id OR
         v_export.delete_lease_until IS DISTINCT FROM
            v_export.maintenance_cleanup_lease_until)) THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export maintenance defer lease is stale';
    END IF;
    UPDATE analytics_dashboard_exports
    SET maintenance_cleanup_lease_owner=NULL,
        maintenance_cleanup_lease_until=NULL,
        delete_lease_owner=CASE WHEN delete_lease_owner=p_cleanup_lease_owner_id
            THEN NULL ELSE delete_lease_owner END,
        delete_lease_until=CASE WHEN delete_lease_owner=p_cleanup_lease_owner_id
            THEN NULL ELSE delete_lease_until END,
        row_revision=row_revision+1,
        updated_at=GREATEST(p_observed_at,updated_at+interval '1 microsecond')
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    RETURN v_retry_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_maintenance_unknown_checkpoint_v1(
    name,text,uuid,name,uuid,text,bigint,bigint,uuid,uuid,integer,bigint,
    uuid,uuid,bytea,timestamptz,bigint,uuid,bigint,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_maintenance_unknown_defer_v1(
    name,text,uuid,name,uuid,text,bigint,bigint,uuid,uuid,integer,bigint,
    uuid,uuid,bytea,timestamptz,bigint,uuid,bigint,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_maintenance_unknown_checkpoint_v1(
    name,text,uuid,name,uuid,text,bigint,bigint,uuid,uuid,integer,bigint,
    uuid,uuid,bytea,timestamptz,bigint,uuid,bigint,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_maintenance_unknown_defer_v1(
    name,text,uuid,name,uuid,text,bigint,bigint,uuid,uuid,integer,bigint,
    uuid,uuid,bytea,timestamptz,bigint,uuid,bigint,timestamptz
) TO forgecrud_runtime;
