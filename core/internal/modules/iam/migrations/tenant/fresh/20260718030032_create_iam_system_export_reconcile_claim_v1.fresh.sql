-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_export_reconcile_claim_v1(
    p_scope_schema name,p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_lease_owner_id uuid,p_lease_generation bigint,
    p_resource_kind text
)
RETURNS TABLE(
    claimed boolean,actor_id uuid,original_job_id uuid,token_version bigint,
    authorization_version bigint,catalog_version_id uuid,catalog_sha256 bytea,
    query_document bytea,query_sha256 bytea,format text,
    original_attempt_generation bigint,original_effect_id uuid,
    original_effect_receipt_id uuid,original_lease_owner_id uuid,
    object_key text,original_request_sha256 bytea,original_started_at timestamptz,
    key_id text,header_sha256 bytea,ciphertext_sha256 bytea,
    ciphertext_bytes bigint,wrapped_dek bytea,wrap_nonce bytea,
    observation_effect_id uuid,observation_receipt_id uuid,
    observation_request_sha256 bytea,observation_started_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_request iam_system_record_exports%ROWTYPE;
    v_effect_id uuid;
    v_receipt_id uuid;
    v_request_sha bytea;
    v_started_at timestamptz;
    v_replayed boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS DISTINCT FROM (
           SELECT identity.tenant_id FROM tenancy_schema_identity AS identity
           WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema()
       ) OR pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_lease_owner_id) IS DISTINCT FROM 7 OR
       p_lease_generation NOT BETWEEN 1 AND 8 OR
       p_resource_kind NOT IN ('users','roles','organizations') THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM export reconciliation claim authority is invalid';
    END IF;
    SELECT request.* INTO v_request FROM iam_system_record_exports AS request
    WHERE request.export_id=p_export_id AND request.tenant_id=p_tenant_id
      AND request.resource_kind=p_resource_kind
      AND request.reconciliation_job_id=p_job_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export reconciliation owner is unavailable';
    END IF;
    IF v_request.reconciliation_state IN ('resolved_sealed','resolved_missing','parked')
       THEN
        RETURN QUERY SELECT false,NULL::uuid,NULL::uuid,NULL::bigint,NULL::bigint,
            NULL::uuid,NULL::bytea,NULL::bytea,NULL::bytea,NULL::text,NULL::bigint,
            NULL::uuid,NULL::uuid,NULL::uuid,NULL::text,NULL::bytea,NULL::timestamptz,
            NULL::text,NULL::bytea,NULL::bytea,NULL::bigint,NULL::bytea,NULL::bytea,
            NULL::uuid,NULL::uuid,NULL::bytea,NULL::timestamptz;
        RETURN;
    END IF;
    IF v_request.state<>'outcome_unknown' OR
       pg_catalog.uuid_extract_version(v_request.effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(v_request.outcome_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(v_request.effect_lease_owner_id) IS DISTINCT FROM 7 OR
       v_request.reconciliation_state NOT IN ('queued','retry_wait','running') OR
       (v_request.reconciliation_attempt_generation IS NOT NULL AND
        v_request.reconciliation_attempt_generation>p_lease_generation) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export reconciliation owner state conflicts';
    END IF;
    IF v_request.reconciliation_state='running' AND
       v_request.reconciliation_attempt_generation=p_lease_generation THEN
        v_effect_id := v_request.reconciliation_effect_id;
        v_receipt_id := v_request.reconciliation_receipt_id;
        v_request_sha := v_request.reconciliation_request_sha256;
        v_started_at := v_request.reconciliation_started_at;
    ELSE
        v_effect_id := p_attempt_id;
        v_receipt_id := pg_catalog.uuidv7();
        v_request_sha := pg_catalog.sha256(
            pg_catalog.convert_to(
                'iam.system-export.reconciliation.v1','UTF8')||
            pg_catalog.decode('00','hex')||pg_catalog.uuid_send(v_request.effect_id)||
            v_request.effect_request_sha256||v_request.ciphertext_sha256||
            v_request.header_sha256||
            pg_catalog.int8send(pg_catalog.octet_length(
                pg_catalog.convert_to(v_request.effect_object_key,'UTF8')))||
            pg_catalog.convert_to(v_request.effect_object_key,'UTF8')||
            pg_catalog.int8send(v_request.ciphertext_bytes));
        v_started_at := pg_catalog.clock_timestamp();
    END IF;
    SELECT platform.platform_async_iam_export_reconcile_effect_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind,v_effect_id,v_receipt_id,
        v_request_sha,v_started_at
    ) INTO v_replayed;
    IF v_replayed IS DISTINCT FROM (
        v_request.reconciliation_state='running' AND
        v_request.reconciliation_attempt_generation=p_lease_generation
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export reconciliation platform replay conflicts';
    END IF;
    UPDATE iam_system_record_exports SET reconciliation_state='running',
        reconciliation_attempt_generation=p_lease_generation,
        reconciliation_effect_id=v_effect_id,reconciliation_receipt_id=v_receipt_id,
        reconciliation_request_sha256=v_request_sha,
        reconciliation_started_at=v_started_at,reconciliation_observed_at=NULL,
        reconciliation_next_at=NULL,reconciliation_code=NULL
    WHERE export_id=p_export_id;
    RETURN QUERY SELECT true,v_request.actor_id,v_request.job_id,
        v_request.token_version,v_request.authorization_version,
        v_request.catalog_version_id,v_request.catalog_sha256,
        v_request.query_document,v_request.query_sha256,v_request.format,
        v_request.attempt_generation,v_request.effect_id,v_request.outcome_receipt_id,
        v_request.effect_lease_owner_id,v_request.effect_object_key,
        v_request.effect_request_sha256,v_request.effect_started_at,v_request.key_id,
        v_request.header_sha256,v_request.ciphertext_sha256,
        v_request.ciphertext_bytes,v_request.wrapped_dek,v_request.wrap_nonce,
        v_effect_id,v_receipt_id,v_request_sha,v_started_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_export_reconcile_claim_v1(
    name,uuid,uuid,uuid,uuid,uuid,bigint,text
) FROM PUBLIC;
