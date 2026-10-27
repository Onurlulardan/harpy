-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_import_source_cleanup_retry_v1(
    p_scope_schema name,p_tenant_id uuid,p_source_id uuid,
    p_lease_owner uuid,p_lease_generation bigint,p_error_code text,
    p_observed_at timestamptz,p_retry_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_source iam_system_import_sources%ROWTYPE;
BEGIN
    SELECT identity.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity
    WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema();
    IF p_tenant_id IS NULL OR
       pg_catalog.uuid_extract_version(p_source_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_lease_owner) IS DISTINCT FROM 7 OR
       p_lease_generation<1 OR p_error_code IS NULL OR
       v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_error_code NOT IN (
           'SOURCE_DELETE_OUTCOME_UNKNOWN','SOURCE_DELETE_UNAVAILABLE'
       ) OR NOT pg_catalog.isfinite(p_observed_at) OR
       NOT pg_catalog.isfinite(p_retry_at) OR p_retry_at<p_observed_at OR
       p_retry_at>p_observed_at+interval '15 minutes' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM import source cleanup retry is invalid';
    END IF;
    SELECT source.* INTO v_source
    FROM iam_system_import_sources AS source
    WHERE source.source_id=p_source_id AND source.tenant_id=p_tenant_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6044',
            MESSAGE='IAM import source was not found';
    END IF;
    IF v_source.state='deleted' THEN
        RETURN true;
    END IF;
    IF v_source.state NOT IN ('ready','claimed') OR
       v_source.cleanup_state<>'running' OR
       v_source.cleanup_lease_owner IS DISTINCT FROM p_lease_owner OR
       v_source.cleanup_lease_generation IS DISTINCT FROM p_lease_generation THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='IAM import source cleanup lease is stale';
    END IF;
    UPDATE iam_system_import_sources
    SET cleanup_state='waiting',cleanup_next_at=p_retry_at,
        cleanup_lease_owner=NULL,cleanup_lease_until=NULL,
        cleanup_last_error_code=p_error_code
    WHERE source_id=p_source_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_import_source_cleanup_retry_v1(
    name,uuid,uuid,uuid,bigint,text,timestamptz,timestamptz
) FROM PUBLIC;
