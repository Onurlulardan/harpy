-- +goose Up
ALTER FUNCTION analytics_dashboard_export_delete_complete_v1(
    name,uuid,uuid,uuid,bigint,text,text,bytea,bytea,timestamptz,timestamptz
) RENAME TO analytics_dashboard_export_delete_complete_unfenced_v1;
REVOKE ALL ON FUNCTION analytics_dashboard_export_delete_complete_unfenced_v1(
    name,uuid,uuid,uuid,bigint,text,text,bytea,bytea,timestamptz,timestamptz
) FROM PUBLIC,forgecrud_runtime;

-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_delete_complete_v1(
    p_scope_schema name,p_export_id uuid,p_delete_job_id uuid,
    p_lease_owner uuid,p_delete_generation bigint,p_object_version text,
    p_provider_request_id text,p_delete_response_sha256 bytea,
    p_absence_sha256 bytea,p_verified_at timestamptz,p_completed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_platform_kind text;
    v_export analytics_dashboard_exports%ROWTYPE;
    v_expected_kind text;
BEGIN
    IF pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_delete_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_lease_owner) IS DISTINCT FROM 5 OR
       p_delete_generation NOT BETWEEN 1 AND 3 OR
       (p_object_version IS NOT NULL AND
        pg_catalog.octet_length(p_object_version) NOT BETWEEN 1 AND 512) OR
       (p_provider_request_id IS NOT NULL AND
        pg_catalog.octet_length(p_provider_request_id) NOT BETWEEN 1 AND 512) OR
       (p_delete_response_sha256 IS NOT NULL AND
        pg_catalog.octet_length(p_delete_response_sha256) IS DISTINCT FROM 32) OR
       (p_absence_sha256 IS NOT NULL AND
        pg_catalog.octet_length(p_absence_sha256) IS DISTINCT FROM 32) OR
       p_verified_at IS NULL OR NOT pg_catalog.isfinite(p_verified_at) OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       p_completed_at<p_verified_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export delete completion input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row
    WHERE identity_row.singleton
      AND identity_row.schema_name=pg_catalog.current_schema();
    v_platform_kind:=
        platform.platform_async_analytics_dashboard_export_maintenance_effect_owner_v1(
            v_tenant_id,p_export_id,p_delete_job_id,p_lease_owner,p_completed_at);
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.tenant_id=v_tenant_id
      AND item.state='deleting' AND item.delete_job_id=p_delete_job_id
      AND item.delete_lease_owner=p_lease_owner
      AND item.delete_attempt_generation=p_delete_generation
    FOR UPDATE;
    v_expected_kind:=CASE WHEN v_export.expiry_job_id=p_delete_job_id THEN
        'analytics.dashboard.export.expire' ELSE
        'analytics.dashboard.export.delete' END;
    IF v_export.export_id IS NULL OR
       v_platform_kind IS DISTINCT FROM v_expected_kind OR
       v_export.maintenance_unknown_job_id IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export delete completion owner changed';
    END IF;
    RETURN analytics_dashboard_export_delete_complete_unfenced_v1(
        p_scope_schema,p_export_id,p_delete_job_id,p_lease_owner,
        p_delete_generation,p_object_version,p_provider_request_id,
        p_delete_response_sha256,p_absence_sha256,p_verified_at,p_completed_at);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_delete_complete_v1(
    name,uuid,uuid,uuid,bigint,text,text,bytea,bytea,timestamptz,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_delete_complete_v1(
    name,uuid,uuid,uuid,bigint,text,text,bytea,bytea,timestamptz,timestamptz
) TO forgecrud_runtime;
