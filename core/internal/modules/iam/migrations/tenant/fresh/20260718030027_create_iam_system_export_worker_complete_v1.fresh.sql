-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_worker_complete_v1(
    p_scope_schema name,p_tenant_id uuid,p_actor_id uuid,p_export_id uuid,
    p_job_id uuid,p_attempt_id uuid,p_lease_owner_id uuid,
    p_lease_generation bigint,p_resource_kind text,p_token_version bigint,
    p_authorization_version bigint,p_processed_rows bigint,
    p_plaintext_bytes bigint,p_checked_at timestamptz,p_object_key text,
    p_object_version text,p_key_id text,p_provider_request_id text,
    p_effect_id uuid,p_effect_receipt_id uuid,p_header_sha256 bytea,
    p_ciphertext_sha256 bytea,p_request_sha256 bytea,p_result_sha256 bytea,
    p_ciphertext_bytes bigint,p_wrapped_dek bytea,p_wrap_nonce bytea,
    p_sealed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_request iam_system_record_exports%ROWTYPE;
    v_replayed boolean;
BEGIN
    IF p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS DISTINCT FROM (
           SELECT identity.tenant_id FROM tenancy_schema_identity AS identity
           WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema()
       ) OR p_processed_rows NOT BETWEEN 0 AND 1000000 OR
       p_plaintext_bytes NOT BETWEEN 0 AND 2147483648 OR
       pg_catalog.octet_length(p_object_key) NOT BETWEEN 1 AND 512 OR
       pg_catalog.octet_length(p_object_version) NOT BETWEEN 1 AND 512 OR
       pg_catalog.octet_length(p_key_id) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_provider_request_id) NOT BETWEEN 1 AND 512 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_receipt_id) IS DISTINCT FROM 7 OR
       p_effect_id=p_effect_receipt_id OR
       pg_catalog.octet_length(p_header_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_ciphertext_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_result_sha256) IS DISTINCT FROM 32 OR
       p_ciphertext_bytes NOT BETWEEN 1 AND 2148532224 OR
       pg_catalog.octet_length(p_wrapped_dek) NOT BETWEEN 32 AND 1024 OR
       pg_catalog.octet_length(p_wrap_nonce) IS DISTINCT FROM 12 OR
       NOT pg_catalog.isfinite(p_checked_at) OR NOT pg_catalog.isfinite(p_sealed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM export completion is invalid';
    END IF;
    SELECT request.* INTO v_request FROM iam_system_record_exports AS request
    WHERE request.export_id=p_export_id AND request.tenant_id=p_tenant_id
      AND request.actor_id=p_actor_id AND request.job_id=p_job_id
      AND request.resource_kind=p_resource_kind
      AND request.token_version=p_token_version
      AND request.authorization_version=p_authorization_version
    FOR UPDATE;
    IF NOT FOUND OR v_request.attempt_generation<>p_lease_generation THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM export completion owner is stale';
    END IF;
    IF v_request.state='ready' THEN
        SELECT platform.platform_async_iam_export_succeed_v1(
            p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
            p_lease_generation,p_resource_kind,p_effect_id,p_effect_receipt_id,
            p_request_sha256,p_result_sha256,p_provider_request_id,p_sealed_at
        ) INTO v_replayed;
        IF v_replayed AND v_request.processed_rows=p_processed_rows AND
           v_request.plaintext_bytes=p_plaintext_bytes AND
           v_request.object_key=p_object_key AND
           v_request.object_version=p_object_version AND v_request.key_id=p_key_id AND
           v_request.provider_request_id=p_provider_request_id AND
           v_request.effect_id=p_effect_id AND
           v_request.effect_receipt_id=p_effect_receipt_id AND
           v_request.header_sha256=p_header_sha256 AND
           v_request.ciphertext_sha256=p_ciphertext_sha256 AND
           v_request.effect_request_sha256=p_request_sha256 AND
           v_request.effect_result_sha256=p_result_sha256 AND
           v_request.ciphertext_bytes=p_ciphertext_bytes AND
           v_request.wrapped_dek=p_wrapped_dek AND
           v_request.wrap_nonce=p_wrap_nonce AND v_request.ready_at=p_sealed_at THEN
            RETURN true;
        END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export completion replay differs';
    END IF;
    IF v_request.state<>'running' OR
       v_request.effect_id IS DISTINCT FROM p_effect_id OR
       v_request.effect_receipt_id IS DISTINCT FROM p_effect_receipt_id OR
       v_request.effect_object_key IS DISTINCT FROM p_object_key OR
       v_request.effect_request_sha256 IS DISTINCT FROM p_request_sha256 OR
       v_request.key_id IS DISTINCT FROM p_key_id OR
       v_request.header_sha256 IS DISTINCT FROM p_header_sha256 OR
       v_request.ciphertext_sha256 IS DISTINCT FROM p_ciphertext_sha256 OR
       v_request.ciphertext_bytes IS DISTINCT FROM p_ciphertext_bytes OR
       v_request.wrapped_dek IS DISTINCT FROM p_wrapped_dek OR
       v_request.wrap_nonce IS DISTINCT FROM p_wrap_nonce OR
       p_processed_rows<v_request.processed_rows OR
       p_plaintext_bytes<v_request.plaintext_bytes OR
       p_checked_at<v_request.started_at OR p_sealed_at<p_checked_at OR
       p_sealed_at<v_request.effect_started_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export completion evidence conflicts';
    END IF;
    IF NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'export',p_token_version,
        p_authorization_version,true
    ) THEN NULL; END IF;
    SELECT platform.platform_async_iam_export_succeed_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind,p_effect_id,p_effect_receipt_id,
        p_request_sha256,p_result_sha256,p_provider_request_id,p_sealed_at
    ) INTO v_replayed;
    IF v_replayed THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export platform completion conflicts';
    END IF;
    UPDATE iam_system_transfer_admissions SET state='released',
        tenant_run_slot=NULL,principal_wait_slot=NULL,released_at=p_sealed_at
    WHERE transfer_kind='export' AND transfer_id=p_export_id
      AND tenant_id=p_tenant_id AND actor_id=p_actor_id AND state='running';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export running admission is unavailable';
    END IF;
    UPDATE iam_system_record_exports SET state='ready',failure_code=NULL,
        processed_rows=p_processed_rows,plaintext_bytes=p_plaintext_bytes,
        object_key=p_object_key,object_version=p_object_version,key_id=p_key_id,
        provider_request_id=p_provider_request_id,header_sha256=p_header_sha256,
        ciphertext_sha256=p_ciphertext_sha256,ciphertext_bytes=p_ciphertext_bytes,
        wrapped_dek=p_wrapped_dek,wrap_nonce=p_wrap_nonce,
        effect_result_sha256=p_result_sha256,effect_completed_at=p_sealed_at,
        ready_at=p_sealed_at,
        expires_at=p_sealed_at+interval '24 hours',completed_at=p_sealed_at
    WHERE export_id=p_export_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_worker_complete_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,bigint,
    bigint,timestamptz,text,text,text,text,uuid,uuid,bytea,bytea,bytea,bytea,
    bigint,bytea,bytea,timestamptz
) FROM PUBLIC;
