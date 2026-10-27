-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_system_record_import_worker_claim_v1(
    p_scope_schema name,p_tenant_id uuid,p_import_id uuid,p_job_id uuid,
    p_attempt_id uuid,p_lease_owner_id uuid,p_lease_generation bigint
)
RETURNS TABLE(
    actor_id uuid,resource_kind text,authorization_version bigint,
    duplicate_action text,mapping_document bytea,next_row_index bigint,
    processed_rows bigint,source_id uuid,source_format text,object_key text,
    object_version text,header_sha256 bytea,ciphertext_sha256 bytea,
    ciphertext_bytes bigint,plaintext_sha256 bytea,plaintext_bytes bigint,
    key_id text,wrapped_dek bytea,wrap_nonce bytea,provider_request_id text,
    uploaded_at timestamptz,finished boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_request iam_system_record_imports%ROWTYPE;
    v_source iam_system_import_sources%ROWTYPE;
    v_cancelled boolean;
    v_slot smallint;
BEGIN
    SELECT identity.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity
    WHERE identity.singleton AND identity.schema_name=pg_catalog.current_schema();
    IF p_scope_schema::text IS DISTINCT FROM pg_catalog.current_schema() OR
       v_tenant_id IS DISTINCT FROM p_tenant_id OR
       pg_catalog.uuid_extract_version(p_import_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_lease_owner_id) IS DISTINCT FROM 7 OR
       p_lease_generation NOT BETWEEN 1 AND 8 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='IAM import worker claim scope is invalid';
    END IF;
    SELECT request.* INTO v_request
    FROM iam_system_record_imports AS request
    WHERE request.import_id=p_import_id AND request.tenant_id=p_tenant_id
      AND request.job_id=p_job_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6044',
            MESSAGE='IAM import worker request was not found';
    END IF;
    SELECT platform.platform_async_iam_import_current_lease_v1(
        p_tenant_id,p_import_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,v_request.resource_kind
    ) INTO v_cancelled;
    IF v_cancelled OR v_request.state='cancelled' THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='IAM import cancellation was requested';
    END IF;
    IF v_request.state NOT IN ('waiting','running','completed') THEN
        RAISE EXCEPTION USING ERRCODE='P6044',
            MESSAGE='IAM import is not executable';
    END IF;
    SELECT source.* INTO STRICT v_source
    FROM iam_system_import_sources AS source
    WHERE source.source_id=v_request.source_id
      AND source.tenant_id=p_tenant_id
      AND source.actor_id=v_request.actor_id
      AND source.resource_kind=v_request.resource_kind
      AND source.claimed_import_id=p_import_id AND source.state='claimed'
    FOR SHARE;
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
                MESSAGE='IAM import tenant run capacity exhausted';
        END IF;
        UPDATE iam_system_transfer_admissions
        SET state='running',tenant_run_slot=v_slot,principal_wait_slot=NULL
        WHERE transfer_kind='import' AND transfer_id=p_import_id
          AND tenant_id=p_tenant_id AND actor_id=v_request.actor_id
          AND state='waiting';
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6044',
                MESSAGE='IAM import admission was not found';
        END IF;
        UPDATE iam_system_record_imports
        SET state='running',started_at=COALESCE(started_at,transaction_timestamp())
        WHERE import_id=p_import_id;
        v_request.state := 'running';
    ELSIF v_request.state='running' AND NOT EXISTS(
        SELECT 1 FROM iam_system_transfer_admissions AS admission
        WHERE admission.transfer_kind='import' AND admission.transfer_id=p_import_id
          AND admission.tenant_id=p_tenant_id AND admission.state='running'
          AND admission.tenant_run_slot BETWEEN 1 AND 2
    ) THEN
        RAISE EXCEPTION USING ERRCODE='P6044',
            MESSAGE='IAM import running admission is invalid';
    END IF;
    RETURN QUERY SELECT v_request.actor_id,v_request.resource_kind,
        v_request.authorization_version,v_request.duplicate_action,
        v_request.mapping_document,v_request.next_row_index,
        v_request.processed_rows,v_source.source_id,v_source.format,
        v_source.object_key,v_source.object_version,v_source.header_sha256,
        v_source.ciphertext_sha256,v_source.ciphertext_bytes,
        v_source.plaintext_sha256,v_source.plaintext_bytes,v_source.key_id,
        v_source.wrapped_dek,v_source.wrap_nonce,v_source.provider_request_id,
        v_source.uploaded_at,v_request.state='completed';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_system_record_import_worker_claim_v1(
    name,uuid,uuid,uuid,uuid,uuid,bigint
) FROM PUBLIC;

