-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_worker_unknown_v1(
    p_scope_schema name,p_tenant_id uuid,p_actor_id uuid,p_export_id uuid,
    p_job_id uuid,p_attempt_id uuid,p_lease_owner_id uuid,
    p_lease_generation bigint,p_resource_kind text,p_token_version bigint,
    p_authorization_version bigint,p_effect_id uuid,p_effect_receipt_id uuid,
    p_object_key text,p_key_id text,p_header_sha256 bytea,
    p_ciphertext_sha256 bytea,p_ciphertext_bytes bigint,p_wrapped_dek bytea,
    p_wrap_nonce bytea,p_request_sha256 bytea,p_started_at timestamptz,
    p_artifact_known boolean,p_object_version text,p_provider_request_id text,
    p_result_sha256 bytea,p_sealed_at timestamptz,p_unknown_at timestamptz,p_code text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_request iam_system_record_exports%ROWTYPE;
    v_replayed boolean;
    v_reconciliation_job_id uuid;
    v_enqueue_replayed boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS DISTINCT FROM (
           SELECT identity.tenant_id FROM tenancy_schema_identity AS identity
           WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema()
       ) OR p_resource_kind NOT IN ('users','roles','organizations') OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_receipt_id) IS DISTINCT FROM 7 OR
       p_effect_id=p_effect_receipt_id OR
       pg_catalog.octet_length(p_object_key) NOT BETWEEN 1 AND 512 OR
       pg_catalog.octet_length(p_key_id) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_header_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_ciphertext_sha256) IS DISTINCT FROM 32 OR
       p_ciphertext_bytes NOT BETWEEN 1 AND 2148532224 OR
       pg_catalog.octet_length(p_wrapped_dek) NOT BETWEEN 32 AND 1024 OR
       pg_catalog.octet_length(p_wrap_nonce) IS DISTINCT FROM 12 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       NOT pg_catalog.isfinite(p_started_at) OR
       NOT pg_catalog.isfinite(p_unknown_at) OR p_unknown_at<p_started_at OR
       p_code<>'JOB_OUTCOME_UNKNOWN' OR p_artifact_known IS NULL OR
       (p_artifact_known AND (
           pg_catalog.octet_length(p_object_version) NOT BETWEEN 1 AND 512 OR
           pg_catalog.octet_length(p_provider_request_id) NOT BETWEEN 1 AND 512 OR
           pg_catalog.octet_length(p_result_sha256) IS DISTINCT FROM 32 OR
           NOT pg_catalog.isfinite(p_sealed_at) OR p_sealed_at<p_started_at OR
           p_unknown_at<p_sealed_at)) OR
       (NOT p_artifact_known AND (p_object_version IS NOT NULL OR
           p_provider_request_id IS NOT NULL OR p_result_sha256 IS NOT NULL OR
           p_sealed_at IS NOT NULL)) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM export unknown authority is invalid';
    END IF;
    SELECT request.* INTO v_request FROM iam_system_record_exports AS request
    WHERE request.export_id=p_export_id AND request.tenant_id=p_tenant_id
      AND request.actor_id=p_actor_id AND request.job_id=p_job_id
      AND request.resource_kind=p_resource_kind
      AND request.token_version=p_token_version
      AND request.authorization_version=p_authorization_version
      AND request.attempt_generation=p_lease_generation
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM export unknown owner is stale';
    END IF;
    SELECT platform.platform_async_iam_export_unknown_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind,p_effect_id,p_request_sha256,p_unknown_at
    ) INTO v_replayed;
    SELECT enqueued.job_id,enqueued.replayed
    INTO v_reconciliation_job_id,v_enqueue_replayed
    FROM platform.platform_async_iam_export_reconcile_enqueue_v1(
        p_tenant_id,p_export_id,p_job_id,p_resource_kind,p_unknown_at
    ) AS enqueued;
    IF v_replayed THEN
        IF v_request.state='outcome_unknown' AND
           v_request.failure_code='JOB_OUTCOME_UNKNOWN' AND
           v_request.effect_id=p_effect_id AND
           v_request.effect_lease_owner_id=p_lease_owner_id AND
           v_request.outcome_receipt_id=p_effect_receipt_id AND
           v_request.effect_object_key=p_object_key AND
           v_request.effect_request_sha256=p_request_sha256 AND
           v_request.effect_started_at=p_started_at AND
           v_request.key_id=p_key_id AND v_request.header_sha256=p_header_sha256 AND
           v_request.ciphertext_sha256=p_ciphertext_sha256 AND
           v_request.ciphertext_bytes=p_ciphertext_bytes AND
           v_request.wrapped_dek=p_wrapped_dek AND v_request.wrap_nonce=p_wrap_nonce AND
           v_request.completed_at=p_unknown_at AND
           v_request.reconciliation_job_id=v_reconciliation_job_id AND
           v_enqueue_replayed AND
           ((NOT p_artifact_known AND v_request.object_key IS NULL AND
             v_request.object_version IS NULL AND v_request.provider_request_id IS NULL AND
             v_request.effect_result_sha256 IS NULL AND
             v_request.effect_completed_at IS NULL) OR
            (p_artifact_known AND v_request.object_key=p_object_key AND
             v_request.object_version=p_object_version AND
             v_request.provider_request_id=p_provider_request_id AND
             v_request.effect_result_sha256=p_result_sha256 AND
             v_request.effect_completed_at=p_sealed_at)) THEN
            RETURN true;
        END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export unknown replay differs';
    END IF;
    IF v_request.state NOT IN ('running','cancelled') OR
       v_request.effect_id IS DISTINCT FROM p_effect_id OR
       v_request.effect_lease_owner_id IS DISTINCT FROM p_lease_owner_id OR
       v_request.effect_receipt_id IS DISTINCT FROM p_effect_receipt_id OR
       v_request.effect_object_key IS DISTINCT FROM p_object_key OR
       v_request.effect_request_sha256 IS DISTINCT FROM p_request_sha256 OR
       v_request.effect_started_at IS DISTINCT FROM p_started_at OR
       v_request.key_id IS DISTINCT FROM p_key_id OR
       v_request.header_sha256 IS DISTINCT FROM p_header_sha256 OR
       v_request.ciphertext_sha256 IS DISTINCT FROM p_ciphertext_sha256 OR
       v_request.ciphertext_bytes IS DISTINCT FROM p_ciphertext_bytes OR
       v_request.wrapped_dek IS DISTINCT FROM p_wrapped_dek OR
       v_request.wrap_nonce IS DISTINCT FROM p_wrap_nonce THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export unknown evidence conflicts';
    END IF;
    UPDATE iam_system_transfer_admissions SET state='released',
        tenant_run_slot=NULL,principal_wait_slot=NULL,released_at=p_unknown_at
    WHERE transfer_kind='export' AND transfer_id=p_export_id
      AND tenant_id=p_tenant_id AND actor_id=p_actor_id AND state='running';
    IF NOT FOUND AND NOT EXISTS(
        SELECT 1 FROM iam_system_transfer_admissions AS admission
        WHERE admission.transfer_kind='export' AND admission.transfer_id=p_export_id
          AND admission.tenant_id=p_tenant_id AND admission.actor_id=p_actor_id
          AND admission.state='released'
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export unknown admission is unavailable';
    END IF;
    UPDATE iam_system_record_exports SET state='outcome_unknown',
        failure_code='JOB_OUTCOME_UNKNOWN',effect_receipt_id=NULL,
        outcome_receipt_id=p_effect_receipt_id,
        object_key=CASE WHEN p_artifact_known THEN p_object_key END,
        object_version=CASE WHEN p_artifact_known THEN p_object_version END,
        provider_request_id=CASE WHEN p_artifact_known THEN p_provider_request_id END,
        effect_result_sha256=CASE WHEN p_artifact_known THEN p_result_sha256 END,
        effect_completed_at=CASE WHEN p_artifact_known THEN p_sealed_at END,
        ready_at=NULL,expires_at=NULL,completed_at=p_unknown_at,
        reconciliation_job_id=v_reconciliation_job_id,
        reconciliation_state='queued',reconciliation_attempt_generation=NULL,
        reconciliation_effect_id=NULL,reconciliation_receipt_id=NULL,
        reconciliation_request_sha256=NULL,reconciliation_started_at=NULL,
        reconciliation_observed_at=NULL,reconciliation_next_at=NULL,
        reconciliation_code=NULL,
        metadata_retain_until=GREATEST(
            metadata_retain_until,p_unknown_at+interval '366 days')
    WHERE export_id=p_export_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_worker_unknown_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,uuid,uuid,
    text,text,bytea,bytea,bigint,bytea,bytea,bytea,timestamptz,boolean,text,
    text,bytea,timestamptz,timestamptz,text
) FROM PUBLIC;
