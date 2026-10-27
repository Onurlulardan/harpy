-- +goose Up
ALTER TABLE platform.analytics_dashboard_export_unknown_resolutions
    ADD COLUMN cleanup_generation bigint,
    ADD COLUMN cleanup_lease_owner_id uuid,
    ADD CONSTRAINT analytics_dashboard_export_unknown_resolution_cleanup_check
    CHECK (
        (cleanup_generation IS NULL AND cleanup_lease_owner_id IS NULL) OR
        (cleanup_generation>0 AND
         pg_catalog.uuid_extract_version(cleanup_lease_owner_id)=7)
    );

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION
platform.platform_async_analytics_dashboard_export_unknown_resolve_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_attempt_generation bigint,p_lease_owner_id uuid,p_effect_id uuid,
    p_request_sha256 bytea,p_resolution_receipt_id uuid,
    p_resolution_kind text,p_resolution_sha256 bytea,p_resolved_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export unknown resolution authority is invalid';
    END IF;
    RAISE EXCEPTION USING ERRCODE='P6047',
        MESSAGE='Analytics dashboard export cleanup lease is required';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_unknown_resolve_v1(
        uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,uuid,text,bytea,timestamptz)
FROM PUBLIC;
