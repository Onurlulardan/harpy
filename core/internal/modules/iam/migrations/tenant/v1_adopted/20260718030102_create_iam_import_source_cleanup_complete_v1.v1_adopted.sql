-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_import_source_cleanup_complete_v1(
    p_scope_schema name,p_tenant_id uuid,p_source_id uuid,p_actor_id uuid,
    p_resource_kind text,p_object_key text,p_object_version text,
    p_lease_owner uuid,p_lease_generation bigint,p_provider_request_id text,
    p_delete_response_sha256 bytea,p_absence_probe_sha256 bytea,
    p_verified_absent_at timestamptz
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
    IF p_tenant_id IS NULL OR p_actor_id IS NULL OR
       pg_catalog.uuid_extract_version(p_source_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_lease_owner) IS DISTINCT FROM 7 OR
       p_lease_generation<1 OR v_tenant_id IS DISTINCT FROM p_tenant_id OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_resource_kind NOT IN ('users','roles','organizations') OR
       p_object_key IS DISTINCT FROM 'iam-system-import-sources/v1/'||
           p_source_id::text||'/1.fcart' OR
       pg_catalog.octet_length(p_object_version) NOT BETWEEN 1 AND 512 OR
       p_provider_request_id IS NULL OR
       pg_catalog.octet_length(p_provider_request_id) NOT BETWEEN 1 AND 256 OR
       p_delete_response_sha256 IS NULL OR p_absence_probe_sha256 IS NULL OR
       pg_catalog.octet_length(p_delete_response_sha256)<>32 OR
       pg_catalog.octet_length(p_absence_probe_sha256)<>32 OR
       NOT pg_catalog.isfinite(p_verified_absent_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM import source cleanup completion is invalid';
    END IF;
    SELECT source.* INTO v_source
    FROM iam_system_import_sources AS source
    WHERE source.source_id=p_source_id AND source.tenant_id=p_tenant_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6044',
            MESSAGE='IAM import source was not found';
    END IF;
    IF v_source.actor_id IS DISTINCT FROM p_actor_id OR
       v_source.resource_kind IS DISTINCT FROM p_resource_kind OR
       v_source.object_key IS DISTINCT FROM p_object_key OR
       v_source.object_version IS DISTINCT FROM p_object_version THEN
        RAISE EXCEPTION USING ERRCODE='P6049',
            MESSAGE='IAM import source cleanup identity differs';
    END IF;
    IF v_source.state='deleted' THEN
        IF v_source.delete_provider_request_id IS DISTINCT FROM p_provider_request_id OR
           v_source.delete_response_sha256 IS DISTINCT FROM p_delete_response_sha256 OR
           v_source.absence_probe_sha256 IS DISTINCT FROM p_absence_probe_sha256 OR
           v_source.verified_absent_at IS DISTINCT FROM p_verified_absent_at THEN
            RAISE EXCEPTION USING ERRCODE='P6049',
                MESSAGE='IAM import source cleanup receipt differs';
        END IF;
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
    SET state='deleted',deleted_at=p_verified_absent_at,
        delete_provider_request_id=p_provider_request_id,
        delete_response_sha256=p_delete_response_sha256,
        absence_probe_sha256=p_absence_probe_sha256,
        verified_absent_at=p_verified_absent_at
    WHERE source_id=p_source_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_import_source_cleanup_complete_v1(
    name,uuid,uuid,uuid,text,text,text,uuid,bigint,text,bytea,bytea,timestamptz
) FROM PUBLIC;
