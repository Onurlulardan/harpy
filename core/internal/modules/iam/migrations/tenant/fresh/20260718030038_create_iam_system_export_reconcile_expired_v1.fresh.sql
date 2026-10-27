-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_reconcile_expired_v1(
    p_scope_schema name,p_tenant_id uuid
)
RETURNS TABLE(recovered boolean,resolution text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_export_id uuid;
    v_request iam_system_record_exports%ROWTYPE;
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_now timestamptz;
    v_recovered_at timestamptz;
    v_retry_at timestamptz;
    v_code text;
    v_current boolean;
    v_replayed boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS DISTINCT FROM (
           SELECT identity.tenant_id FROM tenancy_schema_identity AS identity
           WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema()
       ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM export expired reconciliation authority is invalid';
    END IF;
    v_now := pg_catalog.clock_timestamp();
    SELECT request.export_id INTO v_export_id
    FROM iam_system_record_exports AS request
    JOIN platform.jobs AS job ON job.id=request.reconciliation_job_id
    JOIN platform.job_attempts AS attempt
      ON attempt.job_id=job.id AND attempt.attempt=job.attempts
     AND attempt.lease_generation=job.lease_generation
     AND attempt.lease_owner_id=job.lease_owner_id
    WHERE request.tenant_id=p_tenant_id AND request.state='outcome_unknown'
      AND request.reconciliation_state='running'
      AND request.reconciliation_attempt_generation=job.lease_generation
      AND request.reconciliation_effect_id=attempt.effect_id
      AND pg_catalog.encode(request.reconciliation_request_sha256,'hex')=
          attempt.effect_request_sha256
      AND request.reconciliation_started_at=attempt.effect_started_at
      AND request.reconciliation_receipt_id IS NOT NULL
      AND job.tenant_id=p_tenant_id AND job.scope_kind='tenant'
      AND job.scope_id=p_tenant_id AND job.authority_group_id='iam_system_resources'
      AND job.authority_domain='iam' AND job.writer_origin='v2'
      AND job.kind='iam.system-resource.export-reconcile'
      AND job.payload_revision=1 AND job.effect_class='external'
      AND job.max_attempts=8 AND job.state='running'
      AND job.lease_until<=v_now AND attempt.finished_at IS NULL
      AND attempt.effect_receipt_id IS NULL
    ORDER BY job.lease_until,request.export_id
    LIMIT 1;
    IF NOT FOUND THEN
        RETURN QUERY SELECT false,NULL::text;
        RETURN;
    END IF;
    SELECT request.* INTO v_request FROM iam_system_record_exports AS request
    WHERE request.export_id=v_export_id AND request.tenant_id=p_tenant_id
      AND request.state='outcome_unknown'
      AND request.reconciliation_state='running' FOR UPDATE;
    IF NOT FOUND THEN
        RETURN QUERY SELECT false,NULL::text;
        RETURN;
    END IF;
    SELECT job.* INTO v_job
    FROM platform.jobs AS job
    WHERE job.id=v_request.reconciliation_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.authority_group_id='iam_system_resources'
      AND job.authority_domain='iam' AND job.writer_origin='v2'
      AND job.kind='iam.system-resource.export-reconcile'
      AND job.payload_revision=1 AND job.effect_class='external'
      AND job.max_attempts=8 AND job.state='running'
      AND job.lease_generation=v_request.reconciliation_attempt_generation
      AND job.lease_until<=pg_catalog.clock_timestamp()
    FOR UPDATE;
    IF NOT FOUND THEN
        RETURN QUERY SELECT false,NULL::text;
        RETURN;
    END IF;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.job_id=v_job.id AND attempt.attempt=v_job.attempts
      AND attempt.lease_generation=v_job.lease_generation
      AND attempt.lease_owner_id=v_job.lease_owner_id
      AND attempt.effect_id=v_request.reconciliation_effect_id
      AND attempt.effect_request_sha256=
          pg_catalog.encode(v_request.reconciliation_request_sha256,'hex')
      AND attempt.effect_started_at=v_request.reconciliation_started_at
      AND attempt.finished_at IS NULL AND attempt.effect_receipt_id IS NULL
    FOR UPDATE;
    IF NOT FOUND THEN
        RETURN QUERY SELECT false,NULL::text;
        RETURN;
    END IF;
    SELECT EXISTS(
        SELECT 1 FROM platform.cutover_groups AS group_row
        JOIN platform.domain_cutovers AS domain_row
          ON domain_row.scope_kind=group_row.scope_kind
         AND domain_row.scope_id=group_row.scope_id
        WHERE group_row.scope_kind=v_job.scope_kind
          AND group_row.scope_id=v_job.scope_id
          AND group_row.group_id=v_job.authority_group_id
          AND group_row.generation=v_job.authority_group_generation
          AND group_row.state='v2' AND group_row.writer='v2'
          AND domain_row.domain=v_job.authority_domain
          AND domain_row.generation=v_job.cutover_generation
          AND domain_row.state='v2' AND domain_row.writer='v2'
    ) INTO v_current;
    v_recovered_at := GREATEST(pg_catalog.clock_timestamp(),
        v_job.updated_at+interval '1 microsecond',v_job.lease_until,
        v_request.reconciliation_started_at);
    IF v_current AND v_job.lease_generation<8 THEN
        v_code := 'IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN';
        v_retry_at := v_recovered_at + LEAST(
            interval '5 seconds' * pg_catalog.power(
                2::double precision,
                (v_job.lease_generation-1)::double precision),
            interval '15 minutes')*pg_catalog.random();
    ELSE
        v_code := 'IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_PARKED';
        v_retry_at := NULL;
    END IF;
    SELECT platform.platform_async_iam_export_reconcile_expired_v1(
        p_tenant_id,v_request.export_id,v_job.id,v_attempt.attempt_id,
        v_job.lease_owner_id,v_job.lease_generation,v_request.resource_kind,
        v_request.reconciliation_effect_id,v_request.reconciliation_receipt_id,
        v_request.reconciliation_request_sha256,
        v_request.reconciliation_started_at,v_recovered_at,v_retry_at,v_code
    ) INTO v_replayed;
    UPDATE iam_system_record_exports SET
        reconciliation_state=CASE WHEN v_retry_at IS NULL
            THEN 'parked' ELSE 'retry_wait' END,
        reconciliation_observed_at=v_recovered_at,
        reconciliation_next_at=v_retry_at,reconciliation_code=v_code,
        metadata_retain_until=GREATEST(
            metadata_retain_until,v_recovered_at+interval '366 days')
    WHERE export_id=v_request.export_id;
    RETURN QUERY SELECT true,CASE WHEN v_retry_at IS NULL
        THEN 'parked' ELSE 'requeued' END;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_reconcile_expired_v1(
    name,uuid
) FROM PUBLIC;
