-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_maintenance_resolution_replay_matches_v1(
    p_stored analytics_dashboard_export_maintenance_unknown_resolutions,
    p_expected jsonb
)
RETURNS boolean
LANGUAGE sql IMMUTABLE SET search_path=pg_catalog
AS $function$
SELECT p_expected IS NOT NULL AND pg_catalog.jsonb_build_array(
    p_stored.tenant_id,p_stored.tenant_schema,p_stored.kind,p_stored.export_id,
    p_stored.authority_group_id,p_stored.authority_group_generation,
    p_stored.cutover_generation,p_stored.job_id,p_stored.attempt_id,
    p_stored.attempt_number,p_stored.lease_generation,
    p_stored.original_lease_owner_id,p_stored.effect_id,
    p_stored.effect_request_sha256,p_stored.original_unknown_at,
    p_stored.cleanup_generation,p_stored.cleanup_lease_owner_id,
    p_stored.delete_attempt_generation,p_stored.resolution_receipt_id,
    p_stored.effect_receipt_id,p_stored.effect_receipt_sha256,
    p_stored.resolution_kind,p_stored.resolution_sha256,p_stored.terminal_state,
    p_stored.local_state_sha256,p_stored.local_state_at,p_stored.object_key,
    p_stored.object_version,p_stored.reconcile_provider_request_id,
    p_stored.reconcile_evidence_sha256,p_stored.upload_completed_at,
    p_stored.reconcile_verified_at,p_stored.delete_provider_request_id,
    p_stored.delete_response_sha256,p_stored.absence_sha256,
    p_stored.delete_verified_at,p_stored.resolved_at
)=p_expected;
$function$;

CREATE FUNCTION analytics_dashboard_export_maintenance_resolution_mode_v1(
    p_export analytics_dashboard_exports,p_kind text,p_terminal_state text,
    p_resolution_kind text,p_object_key text,p_object_version text,
    p_delete_request_id text,p_delete_response_sha256 bytea,
    p_absence_sha256 bytea,p_delete_verified_at timestamptz,
    p_unknown_at timestamptz,p_local_state_at timestamptz,
    p_cleanup_lease_owner_id uuid,p_delete_attempt_generation bigint
)
RETURNS text
LANGUAGE sql IMMUTABLE SET search_path=pg_catalog
AS $function$
SELECT CASE
    WHEN p_export.state='deleting' AND p_export.quota_state='used' AND
         p_delete_attempt_generation>=1 AND p_terminal_state='deleted' AND
         p_resolution_kind='deleted' AND p_object_key=p_export.object_key AND
         p_object_version=p_export.object_version AND
         p_export.delete_lease_owner=p_cleanup_lease_owner_id AND
         p_export.delete_lease_until=p_export.maintenance_cleanup_lease_until AND
         p_delete_verified_at>=p_local_state_at AND
         p_delete_verified_at>=p_unknown_at AND p_export.ciphertext_bytes IS NOT NULL
        THEN 'active'
    WHEN p_export.state='deleted' AND p_export.quota_state='released' AND
         p_terminal_state='deleted' AND p_object_key IS NOT DISTINCT FROM p_export.object_key AND
         p_object_version IS NOT DISTINCT FROM p_export.object_version AND
         p_delete_request_id IS NOT DISTINCT FROM p_export.delete_provider_request_id AND
         p_delete_response_sha256 IS NOT DISTINCT FROM p_export.delete_response_sha256 AND
         p_absence_sha256 IS NOT DISTINCT FROM p_export.delete_absence_sha256 AND
         p_delete_verified_at IS NOT DISTINCT FROM p_export.delete_verified_at
        THEN 'completed'
    WHEN p_kind='analytics.dashboard.export.expire' AND
         p_export.state=p_terminal_state AND
         p_export.state IN ('expired','cancelled','failed') AND
         p_export.quota_state='released' AND p_object_key IS NULL AND
         p_resolution_kind='absent' THEN 'terminal'
    ELSE NULL
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    analytics_dashboard_export_maintenance_resolution_replay_matches_v1(
        analytics_dashboard_export_maintenance_unknown_resolutions,jsonb
    )
FROM PUBLIC;
REVOKE ALL ON FUNCTION analytics_dashboard_export_maintenance_resolution_mode_v1(
    analytics_dashboard_exports,text,text,text,text,text,text,bytea,bytea,
    timestamptz,timestamptz,timestamptz,uuid,bigint
) FROM PUBLIC;
