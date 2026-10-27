-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_worker_checkpoint_v1(
    p_scope_schema name,p_tenant_id uuid,p_actor_id uuid,p_export_id uuid,
    p_job_id uuid,p_attempt_id uuid,p_lease_owner_id uuid,
    p_lease_generation bigint,p_resource_kind text,p_token_version bigint,
    p_authorization_version bigint,p_processed_rows bigint,
    p_plaintext_bytes bigint,p_checked_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_request iam_system_record_exports%ROWTYPE;
    v_cancelled boolean;
BEGIN
    IF NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'export',p_token_version,
        p_authorization_version,true
    ) THEN NULL; END IF;
    IF p_tenant_id IS DISTINCT FROM (
        SELECT identity.tenant_id FROM tenancy_schema_identity AS identity
        WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema()
    ) OR p_processed_rows NOT BETWEEN 0 AND 1000000 OR
       p_plaintext_bytes NOT BETWEEN 0 AND 2147483648 OR
       NOT pg_catalog.isfinite(p_checked_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM export checkpoint is invalid';
    END IF;
    SELECT platform.platform_async_iam_export_current_lease_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind
    ) INTO v_cancelled;
    IF v_cancelled THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='IAM export cancellation was requested';
    END IF;
    SELECT request.* INTO v_request FROM iam_system_record_exports AS request
    WHERE request.export_id=p_export_id AND request.tenant_id=p_tenant_id
      AND request.actor_id=p_actor_id AND request.job_id=p_job_id
      AND request.resource_kind=p_resource_kind
      AND request.token_version=p_token_version
      AND request.authorization_version=p_authorization_version
    FOR UPDATE;
    IF NOT FOUND OR v_request.state<>'running' OR
       v_request.attempt_generation<>p_lease_generation OR
       p_processed_rows<v_request.processed_rows OR
       p_plaintext_bytes<v_request.plaintext_bytes OR
       p_checked_at<v_request.started_at THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM export checkpoint lease is stale';
    END IF;
    UPDATE iam_system_record_exports SET processed_rows=p_processed_rows,
        plaintext_bytes=p_plaintext_bytes WHERE export_id=p_export_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_worker_checkpoint_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,
    bigint,bigint,timestamptz
) FROM PUBLIC;
