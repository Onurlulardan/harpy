-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_reconcile_retry_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_lease_owner_id uuid,p_lease_generation bigint,
    p_resource_kind text,p_observation_effect_id uuid,
    p_observation_receipt_id uuid,p_observation_request_sha256 bytea,
    p_observation_started_at timestamptz,p_failure_sha256 bytea,
    p_observed_at timestamptz,p_retry_at timestamptz,p_code text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_request iam_system_record_exports%ROWTYPE;
    v_expected_failure bytea;
    v_replayed boolean;
BEGIN
    v_expected_failure := pg_catalog.sha256(
        pg_catalog.convert_to(
            'iam.system-export.reconciliation-failure.v1','UTF8') ||
        pg_catalog.decode('00','hex') || pg_catalog.uuid_send(p_job_id) ||
        pg_catalog.uuid_send(p_observation_effect_id) ||
        pg_catalog.int8send(pg_catalog.octet_length(
            pg_catalog.convert_to(p_code,'UTF8'))::bigint) ||
        pg_catalog.convert_to(p_code,'UTF8') ||
        pg_catalog.int8send(p_lease_generation));
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS DISTINCT FROM (
           SELECT identity.tenant_id FROM tenancy_schema_identity AS identity
           WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema()
       ) OR p_resource_kind NOT IN ('users','roles','organizations') OR
       p_lease_generation NOT BETWEEN 1 AND 7 OR
       pg_catalog.uuid_extract_version(p_observation_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_observation_receipt_id) IS DISTINCT FROM 7 OR
       p_observation_effect_id=p_observation_receipt_id OR
       pg_catalog.octet_length(p_observation_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_failure_sha256) IS DISTINCT FROM 32 OR
       p_failure_sha256<>v_expected_failure OR
       NOT pg_catalog.isfinite(p_observation_started_at) OR
       NOT pg_catalog.isfinite(p_observed_at) OR
       NOT pg_catalog.isfinite(p_retry_at) OR
       p_observed_at<p_observation_started_at OR p_retry_at<p_observed_at OR
       p_retry_at>p_observed_at+interval '15 minutes' OR
       p_code<>'IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM export reconciliation retry authority is invalid';
    END IF;
    SELECT request.* INTO v_request FROM iam_system_record_exports AS request
    WHERE request.export_id=p_export_id AND request.tenant_id=p_tenant_id
      AND request.resource_kind=p_resource_kind
      AND request.reconciliation_job_id=p_job_id
      AND request.reconciliation_attempt_generation=p_lease_generation
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM export reconciliation retry owner is stale';
    END IF;
    SELECT platform.platform_async_iam_export_reconcile_retry_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind,p_observation_effect_id,
        p_observation_receipt_id,p_observation_request_sha256,p_failure_sha256,
        p_observed_at,p_retry_at
    ) INTO v_replayed;
    IF v_replayed THEN
        IF v_request.state='outcome_unknown' AND
           v_request.reconciliation_state='retry_wait' AND
           v_request.reconciliation_effect_id=p_observation_effect_id AND
           v_request.reconciliation_receipt_id=p_observation_receipt_id AND
           v_request.reconciliation_request_sha256=p_observation_request_sha256 AND
           v_request.reconciliation_started_at=p_observation_started_at AND
           v_request.reconciliation_observed_at=p_observed_at AND
           v_request.reconciliation_next_at=p_retry_at AND
           v_request.reconciliation_code=p_code THEN RETURN true; END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export reconciliation retry replay differs';
    END IF;
    IF v_request.state<>'outcome_unknown' OR
       v_request.reconciliation_state<>'running' OR
       v_request.reconciliation_effect_id<>p_observation_effect_id OR
       v_request.reconciliation_receipt_id<>p_observation_receipt_id OR
       v_request.reconciliation_request_sha256<>p_observation_request_sha256 OR
       v_request.reconciliation_started_at<>p_observation_started_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export reconciliation retry evidence conflicts';
    END IF;
    UPDATE iam_system_record_exports SET reconciliation_state='retry_wait',
        reconciliation_observed_at=p_observed_at,
        reconciliation_next_at=p_retry_at,reconciliation_code=p_code,
        metadata_retain_until=GREATEST(
            metadata_retain_until,p_observed_at+interval '366 days')
    WHERE export_id=p_export_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_reconcile_retry_v1(
    name,uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,timestamptz,
    bytea,timestamptz,timestamptz,text
) FROM PUBLIC;
