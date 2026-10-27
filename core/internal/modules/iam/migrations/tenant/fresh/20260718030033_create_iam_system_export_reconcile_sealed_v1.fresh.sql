-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_reconcile_sealed_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_lease_owner_id uuid,p_lease_generation bigint,
    p_resource_kind text,p_observation_effect_id uuid,
    p_observation_receipt_id uuid,p_observation_request_sha256 bytea,
    p_observation_started_at timestamptz,p_object_key text,p_object_version text,
    p_key_id text,p_artifact_provider_request_id text,p_original_effect_id uuid,
    p_original_receipt_id uuid,p_header_sha256 bytea,p_ciphertext_sha256 bytea,
    p_original_request_sha256 bytea,p_artifact_result_sha256 bytea,
    p_ciphertext_bytes bigint,p_wrapped_dek bytea,p_wrap_nonce bytea,
    p_original_attempt_generation bigint,p_artifact_sealed_at timestamptz,
    p_observation_provider_request_id text,p_observation_result_sha256 bytea,
    p_checked_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_request iam_system_record_exports%ROWTYPE;
    v_replayed boolean;
    v_ready_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS DISTINCT FROM (
           SELECT identity.tenant_id FROM tenancy_schema_identity AS identity
           WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema()
       ) OR p_resource_kind NOT IN ('users','roles','organizations') OR
       pg_catalog.uuid_extract_version(p_observation_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_observation_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_original_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_original_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.octet_length(p_observation_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_object_key) NOT BETWEEN 1 AND 512 OR
       pg_catalog.octet_length(p_object_version) NOT BETWEEN 1 AND 512 OR
       pg_catalog.octet_length(p_key_id) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_artifact_provider_request_id) NOT BETWEEN 1 AND 512 OR
       pg_catalog.octet_length(p_header_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_ciphertext_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_original_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_artifact_result_sha256) IS DISTINCT FROM 32 OR
       p_ciphertext_bytes NOT BETWEEN 1 AND 2148532224 OR
       pg_catalog.octet_length(p_wrapped_dek) NOT BETWEEN 32 AND 1024 OR
       pg_catalog.octet_length(p_wrap_nonce) IS DISTINCT FROM 12 OR
       pg_catalog.octet_length(p_observation_provider_request_id) NOT BETWEEN 1 AND 512 OR
       pg_catalog.octet_length(p_observation_result_sha256) IS DISTINCT FROM 32 OR
       p_original_attempt_generation<1 OR
       NOT pg_catalog.isfinite(p_observation_started_at) OR
       NOT pg_catalog.isfinite(p_artifact_sealed_at) OR
       NOT pg_catalog.isfinite(p_checked_at) OR p_checked_at<p_artifact_sealed_at THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM export sealed reconciliation authority is invalid';
    END IF;
    SELECT request.* INTO v_request FROM iam_system_record_exports AS request
    WHERE request.export_id=p_export_id AND request.tenant_id=p_tenant_id
      AND request.resource_kind=p_resource_kind
      AND request.reconciliation_job_id=p_job_id
      AND request.reconciliation_attempt_generation=p_lease_generation
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM export sealed reconciliation owner is stale';
    END IF;
    SELECT platform.platform_async_iam_export_reconcile_sealed_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind,p_observation_effect_id,
        p_observation_receipt_id,p_observation_request_sha256,
        p_observation_result_sha256,p_observation_provider_request_id,p_checked_at,
        v_request.job_id,p_original_effect_id,p_original_receipt_id,
        p_original_request_sha256,p_artifact_result_sha256,
        p_artifact_provider_request_id,p_artifact_sealed_at
    ) INTO v_replayed;
    IF v_replayed THEN
        IF v_request.state='ready' AND
           v_request.reconciliation_state='resolved_sealed' AND
           v_request.object_key=p_object_key AND
           v_request.object_version=p_object_version AND v_request.key_id=p_key_id AND
           v_request.provider_request_id=p_artifact_provider_request_id AND
           v_request.effect_id=p_original_effect_id AND
           v_request.effect_receipt_id=p_original_receipt_id AND
           v_request.header_sha256=p_header_sha256 AND
           v_request.ciphertext_sha256=p_ciphertext_sha256 AND
           v_request.effect_request_sha256=p_original_request_sha256 AND
           v_request.effect_result_sha256=p_artifact_result_sha256 AND
           v_request.ciphertext_bytes=p_ciphertext_bytes AND
           v_request.wrapped_dek=p_wrapped_dek AND v_request.wrap_nonce=p_wrap_nonce AND
           v_request.attempt_generation=p_original_attempt_generation AND
           v_request.effect_completed_at=p_artifact_sealed_at AND
           v_request.reconciliation_observed_at=p_checked_at THEN RETURN true; END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export sealed owner replay differs';
    END IF;
    IF v_request.state<>'outcome_unknown' OR
       v_request.reconciliation_state<>'running' OR
       v_request.reconciliation_effect_id<>p_observation_effect_id OR
       v_request.reconciliation_receipt_id<>p_observation_receipt_id OR
       v_request.reconciliation_request_sha256<>p_observation_request_sha256 OR
       v_request.reconciliation_started_at<>p_observation_started_at OR
       v_request.effect_id<>p_original_effect_id OR
       v_request.outcome_receipt_id<>p_original_receipt_id OR
       v_request.effect_object_key<>p_object_key OR
       v_request.effect_request_sha256<>p_original_request_sha256 OR
       v_request.key_id<>p_key_id OR v_request.header_sha256<>p_header_sha256 OR
       v_request.ciphertext_sha256<>p_ciphertext_sha256 OR
       v_request.ciphertext_bytes<>p_ciphertext_bytes OR
       v_request.wrapped_dek<>p_wrapped_dek OR v_request.wrap_nonce<>p_wrap_nonce OR
       v_request.attempt_generation<>p_original_attempt_generation OR
       p_checked_at<p_observation_started_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export sealed owner evidence conflicts';
    END IF;
    v_ready_at := GREATEST(p_artifact_sealed_at,v_request.started_at,
        v_request.effect_started_at);
    UPDATE iam_system_record_exports SET state='ready',failure_code=NULL,
        object_key=p_object_key,object_version=p_object_version,key_id=p_key_id,
        provider_request_id=p_artifact_provider_request_id,
        header_sha256=p_header_sha256,ciphertext_sha256=p_ciphertext_sha256,
        ciphertext_bytes=p_ciphertext_bytes,wrapped_dek=p_wrapped_dek,
        wrap_nonce=p_wrap_nonce,effect_receipt_id=p_original_receipt_id,
        outcome_receipt_id=NULL,effect_result_sha256=p_artifact_result_sha256,
        effect_completed_at=p_artifact_sealed_at,ready_at=v_ready_at,
        expires_at=v_ready_at+interval '24 hours',completed_at=p_checked_at,
        reconciliation_state='resolved_sealed',
        reconciliation_observed_at=p_checked_at,reconciliation_next_at=NULL,
        reconciliation_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILED',
        metadata_retain_until=GREATEST(
            metadata_retain_until,p_checked_at+interval '91 days')
    WHERE export_id=p_export_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_reconcile_sealed_v1(
    name,uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,timestamptz,
    text,text,text,text,uuid,uuid,bytea,bytea,bytea,bytea,bigint,bytea,bytea,
    bigint,timestamptz,text,bytea,timestamptz
) FROM PUBLIC;
