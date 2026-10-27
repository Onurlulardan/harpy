-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_import_worker_commit_v1(
    p_scope_schema name,p_tenant_id uuid,p_actor_id uuid,p_import_id uuid,
    p_job_id uuid,p_attempt_id uuid,p_lease_owner_id uuid,
    p_lease_generation bigint,p_resource_kind text,p_token_version bigint,
    p_authorization_version bigint,p_receipts jsonb,p_source_bytes bigint
)
RETURNS TABLE(
    next_row_index bigint,processed_rows bigint,created_rows bigint,
    updated_rows bigint,skipped_rows bigint,rejected_rows bigint,
    decompressed_bytes bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_request iam_system_record_imports%ROWTYPE;
    v_cancelled boolean;
    v_count integer;
    v_created integer;
    v_updated integer;
    v_skipped integer;
    v_rejected integer;
BEGIN
    IF NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'import',p_token_version,
        p_authorization_version,true
    ) THEN NULL; END IF;
    IF p_tenant_id IS DISTINCT FROM (
        SELECT identity.tenant_id FROM tenancy_schema_identity AS identity
        WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema()
    ) OR pg_catalog.jsonb_typeof(p_receipts) IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_array_length(p_receipts) NOT BETWEEN 1 AND 1000 OR
       pg_catalog.octet_length(p_receipts::text) > 1048576 OR
       p_source_bytes NOT BETWEEN 2 AND 1073741824 OR EXISTS(
           SELECT 1 FROM pg_catalog.jsonb_array_elements(p_receipts) AS item
           WHERE pg_catalog.jsonb_typeof(item) IS DISTINCT FROM 'object'
              OR (SELECT pg_catalog.count(*)
                  FROM pg_catalog.jsonb_object_keys(item)) <> 8
              OR NOT item ?& ARRAY['row_index','outcome','owner_kind','owner_id',
                  'owner_revision','error_code','source_sha256','result_sha256']
       ) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM import worker commit document is invalid';
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
      AND request.actor_id=p_actor_id AND request.job_id=p_job_id
      AND request.resource_kind=p_resource_kind
      AND request.authorization_version=p_authorization_version
    FOR UPDATE;
    IF NOT FOUND OR v_request.state IS DISTINCT FROM 'running' THEN
        RAISE EXCEPTION USING ERRCODE='P6044',
            MESSAGE='IAM import worker request is not running';
    END IF;
    WITH decoded AS (
        SELECT * FROM pg_catalog.jsonb_to_recordset(p_receipts) AS receipt(
            row_index bigint,outcome text,owner_kind text,owner_id uuid,
            owner_revision bigint,error_code text,source_sha256 text,
            result_sha256 text
        )
    )
    SELECT pg_catalog.count(*)::integer,
        pg_catalog.count(*) FILTER(WHERE outcome='created')::integer,
        pg_catalog.count(*) FILTER(WHERE outcome='updated')::integer,
        pg_catalog.count(*) FILTER(WHERE outcome='skipped')::integer,
        pg_catalog.count(*) FILTER(WHERE outcome='rejected')::integer
    INTO v_count,v_created,v_updated,v_skipped,v_rejected
    FROM decoded
    HAVING pg_catalog.count(*)=pg_catalog.count(DISTINCT row_index)
       AND pg_catalog.min(row_index)=v_request.next_row_index
       AND pg_catalog.max(row_index)=v_request.next_row_index+
           pg_catalog.count(*)-1
       AND pg_catalog.bool_and(row_index BETWEEN 1 AND 1000000)
       AND pg_catalog.bool_and(source_sha256 ~ '^[0-9a-f]{64}$')
       AND pg_catalog.bool_and(result_sha256 ~ '^[0-9a-f]{64}$')
       AND pg_catalog.bool_and(
           (outcome IN ('created','updated') AND
            owner_kind IN ('record','invitation') AND owner_id IS NOT NULL AND
            owner_revision BETWEEN 1 AND 9007199254740991 AND error_code IS NULL)
           OR (outcome='skipped' AND owner_kind IS NULL AND owner_id IS NULL AND
               owner_revision IS NULL AND error_code IS NULL)
           OR (outcome='rejected' AND owner_kind IS NULL AND owner_id IS NULL AND
               owner_revision IS NULL AND error_code ~ '^[A-Z][A-Z0-9_]{1,63}$')
       )
       AND pg_catalog.bool_and(result_sha256=pg_catalog.encode(
           pg_catalog.sha256(pg_catalog.convert_to(
               'iam.system-transfer.receipt-result.v1'||pg_catalog.chr(0)||
               outcome||pg_catalog.chr(0)||COALESCE(owner_kind,'')||
               pg_catalog.chr(0)||COALESCE(owner_id::text,
                   '00000000-0000-0000-0000-000000000000')||
               pg_catalog.chr(0)||COALESCE(owner_revision,0)::text||
               pg_catalog.chr(0)||COALESCE(error_code,''),'UTF8'
           )),'hex'
       ));
    IF v_count IS NULL OR v_request.processed_rows+v_count>1000000 OR
       v_request.rejected_rows+v_rejected>10000 OR
       v_request.decompressed_bytes+p_source_bytes>2147483648 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM import worker commit exceeds bounded progress';
    END IF;
    INSERT INTO iam_system_import_row_receipts(
        import_id,job_id,row_index,outcome,owner_kind,owner_id,error_code,
        source_sha256,result_sha256,owner_revision,committed_at
    )
    SELECT p_import_id,p_job_id,receipt.row_index,receipt.outcome,
        receipt.owner_kind,receipt.owner_id,receipt.error_code,
        pg_catalog.decode(receipt.source_sha256,'hex'),
        pg_catalog.decode(receipt.result_sha256,'hex'),receipt.owner_revision,
        pg_catalog.transaction_timestamp()
    FROM pg_catalog.jsonb_to_recordset(p_receipts) AS receipt(
        row_index bigint,outcome text,owner_kind text,owner_id uuid,
        owner_revision bigint,error_code text,source_sha256 text,result_sha256 text
    ) ORDER BY receipt.row_index;
    UPDATE iam_system_record_imports AS request
    SET next_row_index=request.next_row_index+v_count,
        processed_rows=request.processed_rows+v_count,
        created_rows=request.created_rows+v_created,
        updated_rows=request.updated_rows+v_updated,
        skipped_rows=request.skipped_rows+v_skipped,
        rejected_rows=request.rejected_rows+v_rejected,
        decompressed_bytes=request.decompressed_bytes+p_source_bytes
    WHERE request.import_id=p_import_id
    RETURNING request.* INTO v_request;
    RETURN QUERY SELECT v_request.next_row_index,v_request.processed_rows,
        v_request.created_rows,v_request.updated_rows,v_request.skipped_rows,
        v_request.rejected_rows,v_request.decompressed_bytes;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_import_worker_commit_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,jsonb,bigint
) FROM PUBLIC;
