-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_worker_finish_v1(
    p_scope_schema name,p_tenant_id uuid,p_actor_id uuid,p_export_id uuid,
    p_job_id uuid,p_attempt_id uuid,p_lease_owner_id uuid,
    p_lease_generation bigint,p_resource_kind text,p_token_version bigint,
    p_authorization_version bigint,p_state text,p_code text,p_finished_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_request iam_system_record_exports%ROWTYPE;
    v_target_state text;
    v_failure_class text;
    v_digest bytea;
    v_replayed boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS DISTINCT FROM (
           SELECT identity.tenant_id FROM tenancy_schema_identity AS identity
           WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema()
       ) OR p_resource_kind NOT IN ('users','roles','organizations') OR
       p_state NOT IN ('failed','cancelled') OR
       p_code !~ '^[A-Z][A-Z0-9_]{1,63}$' OR
       NOT pg_catalog.isfinite(p_finished_at) OR
       (p_state='cancelled' AND p_code<>'IAM_SYSTEM_EXPORT_CANCELLED') THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM export finish authority is invalid';
    END IF;
    v_target_state := p_state;
    v_failure_class := CASE
        WHEN p_state='cancelled' THEN 'cancelled'
        WHEN p_code='IAM_SYSTEM_EXPORT_DEPENDENCY_UNAVAILABLE' THEN 'transient'
        WHEN p_code='IAM_SYSTEM_EXPORT_STATE_CONFLICT' THEN 'conflict'
        ELSE 'permanent'
    END;
    IF v_failure_class='transient' AND p_lease_generation<>3 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM export transient failure is not terminal';
    END IF;
    v_digest := pg_catalog.sha256(pg_catalog.convert_to(p_code,'UTF8'));
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
            MESSAGE='IAM export finish owner is stale';
    END IF;
    SELECT platform.platform_async_iam_export_finish_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind,v_failure_class,p_code,v_digest,
        p_finished_at
    ) INTO v_replayed;
    IF v_replayed THEN
        IF v_request.state=v_target_state AND v_request.failure_code=p_code AND
           v_request.completed_at=p_finished_at AND
           v_request.effect_receipt_id IS NULL THEN
            RETURN true;
        END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export finish replay differs';
    END IF;
    IF (v_target_state='failed' AND v_request.state<>'running') OR
       (v_target_state='cancelled' AND v_request.state NOT IN ('running','cancelled')) OR
       p_finished_at<v_request.started_at OR
       (v_request.effect_started_at IS NOT NULL AND
        p_finished_at<v_request.effect_started_at) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export finish state conflicts';
    END IF;
    UPDATE iam_system_transfer_admissions SET state='released',
        tenant_run_slot=NULL,principal_wait_slot=NULL,released_at=p_finished_at
    WHERE transfer_kind='export' AND transfer_id=p_export_id
      AND tenant_id=p_tenant_id AND actor_id=p_actor_id AND state='running';
    IF NOT FOUND AND NOT EXISTS(
        SELECT 1 FROM iam_system_transfer_admissions AS admission
        WHERE admission.transfer_kind='export' AND admission.transfer_id=p_export_id
          AND admission.tenant_id=p_tenant_id AND admission.actor_id=p_actor_id
          AND admission.state='released'
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export finish admission is unavailable';
    END IF;
    UPDATE iam_system_record_exports SET state=v_target_state,
        failure_code=p_code,object_key=NULL,object_version=NULL,
        header_sha256=NULL,ciphertext_sha256=NULL,ciphertext_bytes=NULL,
        key_id=NULL,wrapped_dek=NULL,wrap_nonce=NULL,effect_receipt_id=NULL,
        provider_request_id=NULL,effect_result_sha256=NULL,
        effect_completed_at=NULL,outcome_receipt_id=NULL,ready_at=NULL,expires_at=NULL,
        completed_at=p_finished_at,metadata_retain_until=GREATEST(
            metadata_retain_until,p_finished_at+CASE WHEN v_target_state='failed'
                THEN interval '366 days' ELSE interval '91 days' END)
    WHERE export_id=p_export_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_worker_finish_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,text,text,
    timestamptz
) FROM PUBLIC;
