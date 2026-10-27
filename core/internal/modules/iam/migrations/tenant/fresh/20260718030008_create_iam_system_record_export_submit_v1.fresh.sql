-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_submit_v1(
    p_scope_schema name,p_export_id uuid,p_job_id uuid,p_receipt_id uuid,
    p_actor_id uuid,p_resource_kind text,p_token_version bigint,
    p_authorization_version bigint,p_authorization_clock bigint,
    p_catalog_version_id uuid,p_catalog_sha256 bytea,p_query_document bytea,
    p_query_sha256 bytea,p_request_sha256 bytea,p_idempotency_sha256 bytea,
    p_format text,p_created_at timestamptz
)
RETURNS TABLE(
    export_id uuid,job_id uuid,resource_kind text,format text,state text,
    processed_rows bigint,plaintext_bytes bigint,created_at timestamptz,
    ready_at timestamptz,expires_at timestamptz,replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_slot smallint;
    v_existing iam_system_record_exports%ROWTYPE;
BEGIN
    IF NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'export',
        p_token_version,p_authorization_version,true
    ) THEN NULL; END IF;
    SELECT identity.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity
    WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema();
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended('iam-transfer-submit:'||p_actor_id::text,6103)
    );
    SELECT request.* INTO v_existing
    FROM iam_system_record_exports AS request
    WHERE request.actor_id=p_actor_id
      AND request.resource_kind=p_resource_kind
      AND request.idempotency_key_sha256=p_idempotency_sha256;
    IF FOUND THEN
        IF v_existing.tenant_id IS DISTINCT FROM v_tenant_id OR
           v_existing.token_version IS DISTINCT FROM p_token_version OR
           v_existing.authorization_version IS DISTINCT FROM p_authorization_version OR
           v_existing.authorization_clock IS DISTINCT FROM p_authorization_clock OR
           v_existing.catalog_version_id IS DISTINCT FROM p_catalog_version_id OR
           v_existing.catalog_sha256 IS DISTINCT FROM p_catalog_sha256 OR
           v_existing.query_document IS DISTINCT FROM p_query_document OR
           v_existing.query_sha256 IS DISTINCT FROM p_query_sha256 OR
           v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 OR
           v_existing.format IS DISTINCT FROM p_format THEN
            RAISE EXCEPTION USING ERRCODE='P6102',
                MESSAGE='IAM export idempotency conflicts';
        END IF;
        RETURN QUERY SELECT v_existing.export_id,v_existing.job_id,
            v_existing.resource_kind,v_existing.format,v_existing.state,
            v_existing.processed_rows,v_existing.plaintext_bytes,
            v_existing.created_at,v_existing.ready_at,v_existing.expires_at,true;
        RETURN;
    END IF;
    SELECT slot::smallint INTO v_slot
    FROM pg_catalog.generate_series(1,5) AS candidate(slot)
    WHERE NOT EXISTS(
        SELECT 1 FROM iam_system_transfer_admissions AS admission
        WHERE admission.actor_id=p_actor_id AND admission.state='waiting'
          AND admission.principal_wait_slot=candidate.slot
    ) ORDER BY slot LIMIT 1;
    IF v_slot IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6101',
            MESSAGE='IAM transfer waiting capacity exhausted';
    END IF;
    INSERT INTO iam_system_record_exports(
        export_id,tenant_id,job_id,idempotency_receipt_id,resource_kind,actor_id,
        token_version,authorization_version,authorization_clock,catalog_version_id,
        catalog_sha256,query_document,query_sha256,request_sha256,
        idempotency_key_sha256,format,created_at,metadata_retain_until
    ) VALUES(
        p_export_id,v_tenant_id,p_job_id,p_receipt_id,p_resource_kind,p_actor_id,
        p_token_version,p_authorization_version,p_authorization_clock,
        p_catalog_version_id,p_catalog_sha256,p_query_document,p_query_sha256,
        p_request_sha256,p_idempotency_sha256,p_format,p_created_at,
        p_created_at+interval '90 days'
    );
    INSERT INTO iam_system_transfer_admissions(
        transfer_kind,transfer_id,tenant_id,actor_id,state,
        principal_wait_slot,admitted_at
    ) VALUES('export',p_export_id,v_tenant_id,p_actor_id,'waiting',v_slot,p_created_at);
    RETURN QUERY SELECT p_export_id,p_job_id,p_resource_kind,p_format,
        'waiting'::text,0::bigint,0::bigint,p_created_at,
        NULL::timestamptz,NULL::timestamptz,false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_submit_v1(
    name,uuid,uuid,uuid,uuid,text,bigint,bigint,bigint,uuid,bytea,bytea,
    bytea,bytea,bytea,text,timestamptz
) FROM PUBLIC;

