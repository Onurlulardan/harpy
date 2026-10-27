-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_unknown_cleanup_checkpoint_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_attempt_generation bigint,p_original_lease_owner uuid,
    p_effect_id uuid,p_effect_request_sha256 bytea,p_cleanup_generation bigint,
    p_cleanup_lease_owner uuid,p_observed_at timestamptz
)
RETURNS TABLE(cleanup_lease_until timestamptz,cancellation_requested boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_tenant_id uuid;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row WHERE identity_row.singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id='00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.uuid_extract_version(p_original_lease_owner) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.octet_length(p_effect_request_sha256) IS DISTINCT FROM 32 OR
       p_cleanup_generation IS NULL OR p_cleanup_generation<1 OR
       pg_catalog.uuid_extract_version(p_cleanup_lease_owner) IS DISTINCT FROM 7 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export cleanup checkpoint scope is invalid';
    END IF;
    RETURN QUERY
    SELECT checkpoint.lease_until,checkpoint.cancellation_requested
    FROM platform.platform_async_analytics_dashboard_export_unknown_cleanup_checkpoint_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_attempt_generation,
        p_original_lease_owner,p_effect_id,p_effect_request_sha256,
        p_cleanup_generation,p_cleanup_lease_owner,p_observed_at
    ) AS checkpoint;
    PERFORM 1 FROM analytics_dashboard_export_unknown_snapshot_v1(
        p_scope_schema,p_tenant_id,p_export_id,p_job_id,p_attempt_id,
        p_attempt_generation,p_original_lease_owner,p_effect_id,
        p_effect_request_sha256
    );
END;
$function$;

CREATE FUNCTION analytics_dashboard_export_unknown_cleanup_defer_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_attempt_generation bigint,p_original_lease_owner uuid,
    p_effect_id uuid,p_effect_request_sha256 bytea,p_cleanup_generation bigint,
    p_cleanup_lease_owner uuid,p_observed_at timestamptz
)
RETURNS timestamptz
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_tenant_id uuid; v_retry_at timestamptz;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT identity_row.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity_row WHERE identity_row.singleton;
    IF v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id='00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.uuid_extract_version(p_original_lease_owner) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.octet_length(p_effect_request_sha256) IS DISTINCT FROM 32 OR
       p_cleanup_generation IS NULL OR p_cleanup_generation<1 OR
       pg_catalog.uuid_extract_version(p_cleanup_lease_owner) IS DISTINCT FROM 7 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export cleanup defer scope is invalid';
    END IF;
    v_retry_at:=platform.platform_async_analytics_dashboard_export_unknown_cleanup_defer_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_attempt_generation,
        p_original_lease_owner,p_effect_id,p_effect_request_sha256,
        p_cleanup_generation,p_cleanup_lease_owner,p_observed_at
    );
    PERFORM 1 FROM analytics_dashboard_export_unknown_snapshot_v1(
        p_scope_schema,p_tenant_id,p_export_id,p_job_id,p_attempt_id,
        p_attempt_generation,p_original_lease_owner,p_effect_id,
        p_effect_request_sha256
    );
    RETURN v_retry_at;
END;
$function$;

CREATE FUNCTION analytics_dashboard_export_unknown_resolution_digest_v1(
    p_kind text,p_object_key text,p_object_version text,
    p_reconcile_request_id text,p_delete_request_id text,
    p_reconcile_sha256 bytea,p_upload_completed_at timestamptz,
    p_reconcile_verified_at timestamptz,p_delete_response_sha256 bytea,
    p_absence_sha256 bytea,p_delete_verified_at timestamptz,
    p_resolved_at timestamptz
)
RETURNS bytea
LANGUAGE sql IMMUTABLE SET search_path FROM CURRENT
AS $function$
SELECT pg_catalog.sha256(
    pg_catalog.convert_to('analytics-dashboard-export-unknown-resolution-v1','UTF8')||
    pg_catalog.int8send(pg_catalog.octet_length(p_kind)::bigint)||
    pg_catalog.convert_to(p_kind,'UTF8')||
    pg_catalog.int8send(pg_catalog.octet_length(p_object_key)::bigint)||
    pg_catalog.convert_to(p_object_key,'UTF8')||
    pg_catalog.int8send(pg_catalog.octet_length(COALESCE(p_object_version,''))::bigint)||
    pg_catalog.convert_to(COALESCE(p_object_version,''),'UTF8')||
    pg_catalog.int8send(pg_catalog.octet_length(p_reconcile_request_id)::bigint)||
    pg_catalog.convert_to(p_reconcile_request_id,'UTF8')||
    pg_catalog.int8send(pg_catalog.octet_length(COALESCE(p_delete_request_id,''))::bigint)||
    pg_catalog.convert_to(COALESCE(p_delete_request_id,''),'UTF8')||
    p_reconcile_sha256||
    pg_catalog.int8send(COALESCE((EXTRACT(EPOCH FROM p_upload_completed_at)*1000000)::bigint,0))||
    pg_catalog.int8send((EXTRACT(EPOCH FROM p_reconcile_verified_at)*1000000)::bigint)||
    COALESCE(p_delete_response_sha256,pg_catalog.decode(pg_catalog.repeat('00',32),'hex'))||
    p_absence_sha256||
    pg_catalog.int8send(COALESCE((EXTRACT(EPOCH FROM p_delete_verified_at)*1000000)::bigint,0))||
    pg_catalog.int8send((EXTRACT(EPOCH FROM p_resolved_at)*1000000)::bigint)
);
$function$;

CREATE FUNCTION analytics_dashboard_export_unknown_resolution_evidence_valid_v1(
    p_kind text,p_summary_sha256 bytea,p_object_key text,p_object_version text,
    p_reconcile_request_id text,p_reconcile_sha256 bytea,
    p_upload_completed_at timestamptz,p_reconcile_verified_at timestamptz,
    p_delete_request_id text,p_delete_response_sha256 bytea,
    p_absence_sha256 bytea,p_delete_verified_at timestamptz,
    p_resolved_at timestamptz
)
RETURNS boolean
LANGUAGE sql IMMUTABLE SET search_path FROM CURRENT
AS $function$
SELECT p_kind IN ('deleted','absent')
   AND pg_catalog.octet_length(p_summary_sha256)=32
   AND p_summary_sha256 IS DISTINCT FROM
       pg_catalog.decode(pg_catalog.repeat('00',32),'hex')
   AND pg_catalog.octet_length(p_object_key) BETWEEN 1 AND 1024
   AND pg_catalog.octet_length(p_reconcile_request_id) BETWEEN 1 AND 512
   AND pg_catalog.octet_length(p_reconcile_sha256)=32
   AND p_reconcile_sha256 IS DISTINCT FROM
       pg_catalog.decode(pg_catalog.repeat('00',32),'hex')
   AND pg_catalog.octet_length(p_absence_sha256)=32
   AND p_absence_sha256 IS DISTINCT FROM
       pg_catalog.decode(pg_catalog.repeat('00',32),'hex')
   AND pg_catalog.isfinite(p_reconcile_verified_at)
   AND pg_catalog.isfinite(p_resolved_at) AND p_reconcile_verified_at<=p_resolved_at
   AND ((p_kind='absent' AND p_object_version IS NULL
        AND p_upload_completed_at IS NULL AND p_delete_request_id IS NULL
        AND p_delete_response_sha256 IS NULL AND p_delete_verified_at IS NULL
        AND p_absence_sha256=p_reconcile_sha256) OR
       (p_kind='deleted' AND pg_catalog.octet_length(p_object_version) BETWEEN 1 AND 512
        AND pg_catalog.octet_length(p_delete_request_id) BETWEEN 1 AND 512
        AND pg_catalog.octet_length(p_delete_response_sha256)=32
        AND p_delete_response_sha256 IS DISTINCT FROM
            pg_catalog.decode(pg_catalog.repeat('00',32),'hex')
        AND pg_catalog.isfinite(p_upload_completed_at)
        AND pg_catalog.isfinite(p_delete_verified_at)
        AND p_upload_completed_at<=p_reconcile_verified_at
        AND p_reconcile_verified_at<=p_delete_verified_at
        AND p_delete_verified_at<=p_resolved_at))
   AND p_summary_sha256=analytics_dashboard_export_unknown_resolution_digest_v1(
        p_kind,p_object_key,p_object_version,p_reconcile_request_id,
        p_delete_request_id,p_reconcile_sha256,p_upload_completed_at,
        p_reconcile_verified_at,p_delete_response_sha256,p_absence_sha256,
        p_delete_verified_at,p_resolved_at);
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_unknown_cleanup_checkpoint_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,bigint,uuid,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_unknown_cleanup_defer_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,bigint,uuid,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_unknown_resolution_digest_v1(
    text,text,text,text,text,bytea,timestamptz,timestamptz,bytea,bytea,
    timestamptz,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_unknown_resolution_evidence_valid_v1(
    text,bytea,text,text,text,bytea,timestamptz,timestamptz,text,bytea,
    bytea,timestamptz,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_unknown_cleanup_checkpoint_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,bigint,uuid,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_unknown_cleanup_defer_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,bigint,uuid,timestamptz
) TO forgecrud_runtime;
