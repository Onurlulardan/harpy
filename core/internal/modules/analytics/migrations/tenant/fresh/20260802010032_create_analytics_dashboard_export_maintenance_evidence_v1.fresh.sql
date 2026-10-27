-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_maintenance_resolution_digest_v1(
    p_kind text,p_terminal_state text,p_object_key text,p_object_version text,
    p_reconcile_request_id text,p_delete_request_id text,
    p_reconcile_sha256 bytea,p_delete_response_sha256 bytea,
    p_absence_sha256 bytea,p_local_state_sha256 bytea,
    p_upload_completed_at timestamptz,p_reconcile_verified_at timestamptz,
    p_delete_verified_at timestamptz,p_local_state_at timestamptz,
    p_resolved_at timestamptz
)
RETURNS bytea
LANGUAGE sql IMMUTABLE SET search_path=pg_catalog
AS $function$
SELECT pg_catalog.sha256(
    pg_catalog.convert_to('analytics-dashboard-export-maintenance-resolution-v1','UTF8')||
    pg_catalog.int8send(pg_catalog.octet_length(p_kind)::bigint)||
    pg_catalog.convert_to(p_kind,'UTF8')||
    pg_catalog.int8send(pg_catalog.octet_length(p_terminal_state)::bigint)||
    pg_catalog.convert_to(p_terminal_state,'UTF8')||
    pg_catalog.int8send(pg_catalog.octet_length(COALESCE(p_object_key,''))::bigint)||
    pg_catalog.convert_to(COALESCE(p_object_key,''),'UTF8')||
    pg_catalog.int8send(pg_catalog.octet_length(COALESCE(p_object_version,''))::bigint)||
    pg_catalog.convert_to(COALESCE(p_object_version,''),'UTF8')||
    pg_catalog.int8send(pg_catalog.octet_length(COALESCE(p_reconcile_request_id,''))::bigint)||
    pg_catalog.convert_to(COALESCE(p_reconcile_request_id,''),'UTF8')||
    pg_catalog.int8send(pg_catalog.octet_length(COALESCE(p_delete_request_id,''))::bigint)||
    pg_catalog.convert_to(COALESCE(p_delete_request_id,''),'UTF8')||
    COALESCE(p_reconcile_sha256,pg_catalog.decode(pg_catalog.repeat('00',32),'hex'))||
    COALESCE(p_delete_response_sha256,pg_catalog.decode(pg_catalog.repeat('00',32),'hex'))||
    COALESCE(p_absence_sha256,pg_catalog.decode(pg_catalog.repeat('00',32),'hex'))||
    p_local_state_sha256||
    pg_catalog.int8send(COALESCE((EXTRACT(EPOCH FROM p_upload_completed_at)*1000000)::bigint,0))||
    pg_catalog.int8send(COALESCE((EXTRACT(EPOCH FROM p_reconcile_verified_at)*1000000)::bigint,0))||
    pg_catalog.int8send(COALESCE((EXTRACT(EPOCH FROM p_delete_verified_at)*1000000)::bigint,0))||
    pg_catalog.int8send((EXTRACT(EPOCH FROM p_local_state_at)*1000000)::bigint)||
    pg_catalog.int8send((EXTRACT(EPOCH FROM p_resolved_at)*1000000)::bigint)
);
$function$;

CREATE FUNCTION analytics_dashboard_export_maintenance_resolution_valid_v1(
    p_kind text,p_resolution_kind text,p_summary_sha256 bytea,
    p_terminal_state text,p_local_state_sha256 bytea,p_local_state_at timestamptz,
    p_object_key text,p_object_version text,p_reconcile_request_id text,
    p_reconcile_sha256 bytea,p_upload_completed_at timestamptz,
    p_reconcile_verified_at timestamptz,p_delete_request_id text,
    p_delete_response_sha256 bytea,p_absence_sha256 bytea,
    p_delete_verified_at timestamptz,p_resolved_at timestamptz
)
RETURNS boolean
LANGUAGE sql IMMUTABLE SET search_path FROM CURRENT
AS $function$
SELECT p_kind IN ('analytics.dashboard.export.delete',
                  'analytics.dashboard.export.expire')
   AND p_resolution_kind IN ('deleted','absent')
   AND pg_catalog.octet_length(p_summary_sha256)=32
   AND pg_catalog.octet_length(p_local_state_sha256)=32
   AND p_summary_sha256 IS DISTINCT FROM
       pg_catalog.decode(pg_catalog.repeat('00',32),'hex')
   AND p_local_state_sha256 IS DISTINCT FROM
       pg_catalog.decode(pg_catalog.repeat('00',32),'hex')
   AND pg_catalog.isfinite(p_local_state_at)
   AND pg_catalog.isfinite(p_resolved_at) AND p_local_state_at<=p_resolved_at
   AND ((p_kind='analytics.dashboard.export.delete' AND p_terminal_state='deleted') OR
        (p_kind='analytics.dashboard.export.expire' AND
         p_terminal_state IN ('expired','cancelled','failed','deleted')))
   AND ((p_object_key IS NULL AND p_object_version IS NULL
         AND p_resolution_kind='absent' AND p_reconcile_request_id IS NULL
         AND p_reconcile_sha256 IS NULL AND p_upload_completed_at IS NULL
         AND p_reconcile_verified_at IS NULL AND p_delete_request_id IS NULL
         AND p_delete_response_sha256 IS NULL AND p_absence_sha256 IS NULL
         AND p_delete_verified_at IS NULL) OR
        (pg_catalog.octet_length(p_object_key) BETWEEN 1 AND 1024
          AND pg_catalog.octet_length(p_object_version) BETWEEN 1 AND 512
          AND pg_catalog.octet_length(p_absence_sha256)=32
          AND p_absence_sha256 IS DISTINCT FROM
              pg_catalog.decode(pg_catalog.repeat('00',32),'hex')
          AND p_resolution_kind='deleted'
                AND pg_catalog.octet_length(p_delete_request_id) BETWEEN 1 AND 512
                AND pg_catalog.octet_length(p_delete_response_sha256)=32
                AND p_delete_response_sha256 IS DISTINCT FROM
                    pg_catalog.decode(pg_catalog.repeat('00',32),'hex')
                AND pg_catalog.isfinite(p_delete_verified_at)
               AND ((p_reconcile_request_id IS NULL
                     AND p_reconcile_sha256 IS NULL
                     AND p_upload_completed_at IS NULL
                     AND p_reconcile_verified_at IS NULL) OR
                     (pg_catalog.octet_length(p_reconcile_request_id) BETWEEN 1 AND 512
                      AND pg_catalog.octet_length(p_reconcile_sha256)=32
                      AND p_reconcile_sha256 IS DISTINCT FROM
                          pg_catalog.decode(pg_catalog.repeat('00',32),'hex')
                      AND (p_upload_completed_at IS NULL OR
                          pg_catalog.isfinite(p_upload_completed_at))
                     AND pg_catalog.isfinite(p_reconcile_verified_at)
                     AND (p_upload_completed_at IS NULL OR
                          p_upload_completed_at<=p_reconcile_verified_at)
                     AND p_reconcile_verified_at<=p_delete_verified_at))
         AND COALESCE(p_reconcile_verified_at,p_delete_verified_at)<=p_resolved_at))
   AND p_summary_sha256=analytics_dashboard_export_maintenance_resolution_digest_v1(
        p_resolution_kind,p_terminal_state,p_object_key,p_object_version,
        p_reconcile_request_id,p_delete_request_id,p_reconcile_sha256,
        p_delete_response_sha256,p_absence_sha256,p_local_state_sha256,
        p_upload_completed_at,p_reconcile_verified_at,p_delete_verified_at,
        p_local_state_at,p_resolved_at);
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_maintenance_resolution_digest_v1(
    text,text,text,text,text,text,bytea,bytea,bytea,bytea,timestamptz,
    timestamptz,timestamptz,timestamptz,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_maintenance_resolution_valid_v1(
    text,text,bytea,text,bytea,timestamptz,text,text,text,bytea,timestamptz,
    timestamptz,text,bytea,bytea,timestamptz,timestamptz
) FROM PUBLIC;
