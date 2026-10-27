-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_import_worker_finish_v1(
    p_scope_schema name,p_tenant_id uuid,p_actor_id uuid,p_import_id uuid,
    p_job_id uuid,p_attempt_id uuid,p_lease_owner_id uuid,
    p_lease_generation bigint,p_resource_kind text,p_token_version bigint,
    p_authorization_version bigint,p_success boolean
)
RETURNS TABLE(
    state text,processed_rows bigint,created_rows bigint,updated_rows bigint,
    skipped_rows bigint,rejected_rows bigint,decompressed_bytes bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_request iam_system_record_imports%ROWTYPE;
    v_cancelled boolean;
    v_target_state text;
BEGIN
    IF p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS DISTINCT FROM (
           SELECT identity.tenant_id FROM tenancy_schema_identity AS identity
           WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema()
       ) OR p_resource_kind NOT IN ('users','roles','organizations') OR
       p_actor_id IS NULL OR p_authorization_version<1 OR p_success IS NULL OR
       (p_success AND p_token_version<1) OR
       (NOT p_success AND p_token_version<>0) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM import worker finish scope is invalid';
    END IF;
    SELECT platform.platform_async_iam_import_current_lease_v1(
        p_tenant_id,p_import_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind
    ) INTO v_cancelled;
    IF v_cancelled THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='IAM import cancellation was requested';
    END IF;
    SELECT request.* INTO v_request
    FROM iam_system_record_imports AS request
    WHERE request.import_id=p_import_id AND request.tenant_id=p_tenant_id
      AND request.job_id=p_job_id AND request.resource_kind=p_resource_kind
      AND request.actor_id=p_actor_id
      AND request.authorization_version=p_authorization_version
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6044',
            MESSAGE='IAM import worker request was not found';
    END IF;
    v_target_state := CASE WHEN p_success THEN 'completed' ELSE 'failed' END;
    IF p_success AND NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'import',p_token_version,
        p_authorization_version,true
    ) THEN NULL; END IF;
    IF v_request.state=v_target_state THEN
        RETURN QUERY SELECT v_request.state,v_request.processed_rows,
            v_request.created_rows,v_request.updated_rows,v_request.skipped_rows,
            v_request.rejected_rows,v_request.decompressed_bytes;
        RETURN;
    END IF;
    IF (p_success AND v_request.state<>'running') OR
       (NOT p_success AND v_request.state NOT IN ('waiting','running')) THEN
        RAISE EXCEPTION USING ERRCODE='P6044',
            MESSAGE='IAM import worker terminal transition is invalid';
    END IF;
    UPDATE iam_system_transfer_admissions AS admission
    SET state='released',tenant_run_slot=NULL,principal_wait_slot=NULL,
        released_at=pg_catalog.transaction_timestamp()
    WHERE admission.transfer_kind='import' AND admission.transfer_id=p_import_id
      AND admission.tenant_id=p_tenant_id AND admission.state IN ('waiting','running');
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6044',
            MESSAGE='IAM import worker admission was not found';
    END IF;
    UPDATE iam_system_record_imports AS request
    SET state=v_target_state,
        completed_at=pg_catalog.transaction_timestamp()
    WHERE request.import_id=p_import_id
    RETURNING request.* INTO v_request;
    RETURN QUERY SELECT v_request.state,v_request.processed_rows,
        v_request.created_rows,v_request.updated_rows,v_request.skipped_rows,
        v_request.rejected_rows,v_request.decompressed_bytes;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_import_worker_finish_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,boolean
) FROM PUBLIC;
