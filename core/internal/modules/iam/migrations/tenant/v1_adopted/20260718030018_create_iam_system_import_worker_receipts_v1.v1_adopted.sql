-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_import_worker_receipts_v1(
    p_scope_schema name,p_tenant_id uuid,p_actor_id uuid,p_import_id uuid,
    p_job_id uuid,p_attempt_id uuid,p_lease_owner_id uuid,
    p_lease_generation bigint,p_resource_kind text,p_token_version bigint,
    p_authorization_version bigint,p_row_indexes bigint[]
)
RETURNS TABLE(
    row_index bigint,outcome text,owner_kind text,owner_id uuid,
    owner_revision bigint,error_code text,source_sha256 bytea,result_sha256 bytea
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_cancelled boolean;
BEGIN
    IF NOT iam_system_transfer_assert_v1(
        p_scope_schema,p_actor_id,p_resource_kind,'import',p_token_version,
        p_authorization_version,true
    ) THEN NULL; END IF;
    IF p_tenant_id IS DISTINCT FROM (
        SELECT identity.tenant_id FROM tenancy_schema_identity AS identity
        WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema()
    ) OR pg_catalog.cardinality(p_row_indexes) NOT BETWEEN 1 AND 1000 OR
       EXISTS(
           SELECT 1 FROM pg_catalog.unnest(p_row_indexes) AS wanted(row_index)
           WHERE wanted.row_index NOT BETWEEN 1 AND 1000000
       ) OR (
           SELECT pg_catalog.count(*) FROM pg_catalog.unnest(p_row_indexes)
       ) IS DISTINCT FROM (
           SELECT pg_catalog.count(DISTINCT wanted.row_index)
           FROM pg_catalog.unnest(p_row_indexes) AS wanted(row_index)
       ) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM import worker receipt request is invalid';
    END IF;
    SELECT platform.platform_async_iam_import_current_lease_v1(
        p_tenant_id,p_import_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind
    ) INTO v_cancelled;
    IF v_cancelled THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='IAM import cancellation was requested';
    END IF;
    IF NOT EXISTS(
        SELECT 1 FROM iam_system_record_imports AS request
        WHERE request.import_id=p_import_id AND request.tenant_id=p_tenant_id
          AND request.actor_id=p_actor_id AND request.job_id=p_job_id
          AND request.resource_kind=p_resource_kind
          AND request.authorization_version=p_authorization_version
          AND request.state IN ('running','completed')
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6044',
            MESSAGE='IAM import worker request is unavailable';
    END IF;
    RETURN QUERY
    SELECT receipt.row_index,receipt.outcome,receipt.owner_kind,
        receipt.owner_id,receipt.owner_revision,receipt.error_code,
        receipt.source_sha256,receipt.result_sha256
    FROM iam_system_import_row_receipts AS receipt
    WHERE receipt.import_id=p_import_id
      AND receipt.job_id=p_job_id
      AND receipt.row_index=ANY(p_row_indexes)
    ORDER BY receipt.row_index;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_import_worker_receipts_v1(
    name,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,bigint,bigint,bigint[]
) FROM PUBLIC;
