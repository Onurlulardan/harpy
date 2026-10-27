-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_worker_claim_v1(
    p_scope_schema name,p_tenant_id uuid,p_actor_id uuid,p_export_id uuid,
    p_job_id uuid,p_attempt_id uuid,p_lease_owner_id uuid,
    p_lease_generation bigint,p_resource_kind text,p_token_version bigint,
    p_authorization_version bigint
)
RETURNS TABLE(state text,replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_request iam_system_record_exports%ROWTYPE;
    v_cancelled boolean;
    v_slot smallint;
BEGIN
    IF NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'export',p_token_version,
        p_authorization_version,true
    ) THEN NULL; END IF;
    IF p_tenant_id IS DISTINCT FROM (
        SELECT identity.tenant_id FROM tenancy_schema_identity AS identity
        WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema()
    ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM export worker tenant is invalid';
    END IF;
    SELECT request.* INTO v_request FROM iam_system_record_exports AS request
    WHERE request.export_id=p_export_id AND request.tenant_id=p_tenant_id
      AND request.actor_id=p_actor_id AND request.job_id=p_job_id
      AND request.resource_kind=p_resource_kind
      AND request.token_version=p_token_version
      AND request.authorization_version=p_authorization_version
    FOR UPDATE;
    IF NOT FOUND OR v_request.state NOT IN ('waiting','running') THEN
        RAISE EXCEPTION USING ERRCODE='P6044',
            MESSAGE='IAM export worker request is unavailable';
    END IF;
    SELECT platform.platform_async_iam_export_current_lease_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind
    ) INTO v_cancelled;
    IF v_cancelled OR v_request.state='cancelled' THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='IAM export cancellation was requested';
    END IF;
    IF v_request.state='running' AND
       v_request.attempt_generation=p_lease_generation THEN
        IF NOT EXISTS(
            SELECT 1 FROM iam_system_transfer_admissions AS admission
            WHERE admission.transfer_kind='export'
              AND admission.transfer_id=p_export_id
              AND admission.tenant_id=p_tenant_id AND admission.state='running'
              AND admission.tenant_run_slot BETWEEN 1 AND 2
        ) THEN
            RAISE EXCEPTION USING ERRCODE='P6044',
                MESSAGE='IAM export running admission is invalid';
        END IF;
        RETURN QUERY SELECT 'running'::text,true;
        RETURN;
    END IF;
    IF v_request.attempt_generation IS NOT NULL AND
       v_request.attempt_generation>=p_lease_generation THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='IAM export attempt generation is stale';
    END IF;
    IF v_request.state='waiting' THEN
        SELECT slot::smallint INTO v_slot
        FROM pg_catalog.generate_series(1,2) AS candidate(slot)
        WHERE NOT EXISTS(
            SELECT 1 FROM iam_system_transfer_admissions AS admission
            WHERE admission.tenant_id=p_tenant_id AND admission.state='running'
              AND admission.tenant_run_slot=candidate.slot
        ) ORDER BY slot LIMIT 1;
        IF v_slot IS NULL THEN
            RAISE EXCEPTION USING ERRCODE='P6101',
                MESSAGE='IAM export tenant run capacity exhausted';
        END IF;
        UPDATE iam_system_transfer_admissions SET state='running',
            tenant_run_slot=v_slot,principal_wait_slot=NULL
        WHERE transfer_kind='export' AND transfer_id=p_export_id
          AND tenant_id=p_tenant_id AND actor_id=p_actor_id AND state='waiting';
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6044',
                MESSAGE='IAM export admission was not found';
        END IF;
    END IF;
    UPDATE iam_system_record_exports SET state='running',
        started_at=COALESCE(started_at,pg_catalog.transaction_timestamp()),
        attempt_generation=p_lease_generation,failure_code=NULL,
        effect_id=NULL,effect_lease_owner_id=NULL,effect_object_key=NULL,
        effect_request_sha256=NULL,
        effect_started_at=NULL,effect_receipt_id=NULL,provider_request_id=NULL,
        effect_result_sha256=NULL,effect_completed_at=NULL,outcome_receipt_id=NULL
    WHERE export_id=p_export_id;
    RETURN QUERY SELECT 'running'::text,false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_worker_claim_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint
) FROM PUBLIC;
