-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_unknown_cleanup_claim_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_attempt_generation bigint,p_original_lease_owner uuid,
    p_effect_id uuid,p_effect_request_sha256 bytea,
    p_cleanup_lease_owner uuid,p_observed_at timestamptz
)
RETURNS TABLE(
    claim_state text,cleanup_generation bigint,
    cleanup_lease_until timestamptz,format text,object_key text,
    key_id text,wrapped_dek bytea,wrap_nonce bytea,header_sha256 bytea,
    ciphertext_sha256 bytea,plaintext_sha256 bytea,ciphertext_bytes bigint,
    plaintext_bytes bigint,row_count bigint,content_type text,filename text,
    effect_started_at timestamptz,cancellation_requested boolean,
    expires_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_claim_state text;
    v_cleanup_generation bigint;
    v_cleanup_lease_until timestamptz;
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
       p_effect_request_sha256=pg_catalog.decode(pg_catalog.repeat('00',32),'hex') OR
       pg_catalog.uuid_extract_version(p_cleanup_lease_owner) IS DISTINCT FROM 7 OR
       p_observed_at IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export cleanup claim scope is invalid';
    END IF;
    SELECT claimed.claim_state,claimed.cleanup_generation,claimed.lease_until
    INTO STRICT v_claim_state,v_cleanup_generation,v_cleanup_lease_until
    FROM platform.platform_async_analytics_dashboard_export_unknown_cleanup_claim_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_attempt_generation,
        p_original_lease_owner,p_effect_id,p_effect_request_sha256,
        p_cleanup_lease_owner,p_observed_at
    ) AS claimed;
    claim_state:=v_claim_state;
    cleanup_generation:=v_cleanup_generation;
    cleanup_lease_until:=v_cleanup_lease_until;
    IF v_claim_state IS DISTINCT FROM 'claimed' THEN
        IF v_claim_state NOT IN ('busy','resolved') OR
           v_cleanup_generation IS NOT NULL OR v_cleanup_lease_until IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export cleanup claim result is invalid';
        END IF;
        RETURN NEXT; RETURN;
    END IF;
    IF v_cleanup_generation IS NULL OR v_cleanup_generation<1 OR
       v_cleanup_lease_until IS NULL OR v_cleanup_lease_until<=p_observed_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export cleanup lease is invalid';
    END IF;
    SELECT snapshot.format,snapshot.object_key,snapshot.key_id,
           snapshot.wrapped_dek,snapshot.wrap_nonce,snapshot.header_sha256,
           snapshot.ciphertext_sha256,snapshot.plaintext_sha256,
           snapshot.ciphertext_bytes,snapshot.plaintext_bytes,
           snapshot.row_count,snapshot.content_type,snapshot.filename,
           snapshot.effect_started_at,snapshot.cancellation_requested,
           snapshot.expires_at
    INTO STRICT format,object_key,key_id,wrapped_dek,wrap_nonce,header_sha256,
         ciphertext_sha256,plaintext_sha256,ciphertext_bytes,plaintext_bytes,
         row_count,content_type,filename,effect_started_at,
         cancellation_requested,expires_at
    FROM analytics_dashboard_export_unknown_snapshot_v1(
        p_scope_schema,p_tenant_id,p_export_id,p_job_id,p_attempt_id,
        p_attempt_generation,p_original_lease_owner,p_effect_id,
        p_effect_request_sha256
    ) AS snapshot;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_unknown_cleanup_claim_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_unknown_cleanup_claim_v1(
    name,uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,uuid,timestamptz
) TO harpy_runtime;
