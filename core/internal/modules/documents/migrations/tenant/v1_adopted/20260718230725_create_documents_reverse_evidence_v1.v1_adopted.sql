-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_v1_reverse_evidence_v1(
    p_reconciliation_id uuid,p_tenant_id uuid,
    p_source_generation bigint,p_target_generation bigint,
    p_source_count bigint,p_target_count bigint,
    p_source_sha256 bytea,p_target_sha256 bytea,
    p_requested_by uuid,p_approved_by uuid,p_accepted_at timestamptz
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_receipt_sha text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_cutover_finalizer' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Documents V1 reverse principal denied';
    END IF;
    IF pg_catalog.uuid_extract_version(p_reconciliation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_requested_by) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_approved_by) IS DISTINCT FROM 7 OR
       p_requested_by=p_approved_by OR p_source_generation<1 OR
       p_target_generation<=p_source_generation OR p_source_count<0 OR
       p_source_count IS DISTINCT FROM p_target_count OR
       pg_catalog.octet_length(p_source_sha256)<>32 OR
       p_source_sha256 IS DISTINCT FROM p_target_sha256 OR
       p_accepted_at IS NULL OR NOT pg_catalog.isfinite(p_accepted_at) OR
       p_accepted_at>pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Documents V1 reverse evidence is invalid';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM tenancy_schema_identity AS identity
        WHERE identity.singleton AND identity.tenant_id=p_tenant_id
          AND identity.schema_name=pg_catalog.current_schema()
          AND identity.origin='v1_adopted'
    ) OR NOT EXISTS (
        SELECT 1 FROM platform.domain_cutovers AS authority
        WHERE authority.scope_kind='tenant' AND authority.scope_id=p_tenant_id
          AND authority.domain='documents'
          AND authority.generation=p_target_generation
          AND authority.state IN ('v2','fenced')
    ) OR NOT EXISTS (
        SELECT 1 FROM documents_v1_cutover_receipts AS receipt
        WHERE receipt.tenant_id=p_tenant_id
          AND receipt.cutover_generation=p_target_generation
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 reverse authority is not ready';
    END IF;
    v_receipt_sha:=pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
        'harpy/documents/v1-reverse-receipt/v1:'||
        p_reconciliation_id::text||':'||p_tenant_id::text||':'||
        p_source_generation::text||':'||p_target_generation::text||':'||
        p_source_count::text||':'||pg_catalog.encode(p_source_sha256,'hex')||':'||
        p_requested_by::text||':'||p_approved_by::text||':'||p_accepted_at::text,
        'UTF8'
    )),'hex');
    RETURN v_receipt_sha;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_v1_reverse_evidence_v1(
    uuid,uuid,bigint,bigint,bigint,bigint,bytea,bytea,uuid,uuid,timestamptz
) FROM PUBLIC;
