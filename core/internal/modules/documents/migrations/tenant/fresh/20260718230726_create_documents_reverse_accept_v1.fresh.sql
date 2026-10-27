-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_v1_reverse_accept_v1(
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
    v_receipt_sha:=documents_v1_reverse_evidence_v1(
        p_reconciliation_id,p_tenant_id,p_source_generation,p_target_generation,
        p_source_count,p_target_count,p_source_sha256,p_target_sha256,
        p_requested_by,p_approved_by,p_accepted_at
    );
    INSERT INTO documents_reverse_reconciliations (
        reconciliation_id,source_generation,target_generation,
        source_count,target_count,source_sha256,target_sha256,lifecycle_state,
        requested_by,approved_by,created_at,decided_at
    ) VALUES (
        p_reconciliation_id,p_source_generation,p_target_generation,
        p_source_count,p_target_count,p_source_sha256,p_target_sha256,'ACCEPTED',
        p_requested_by,p_approved_by,p_accepted_at,p_accepted_at
    ) ON CONFLICT (reconciliation_id) DO NOTHING;
    IF NOT EXISTS (
        SELECT 1 FROM documents_reverse_reconciliations AS reconciliation
        WHERE reconciliation.reconciliation_id=p_reconciliation_id
          AND reconciliation.source_generation=p_source_generation
          AND reconciliation.target_generation=p_target_generation
          AND reconciliation.source_count=p_source_count
          AND reconciliation.target_count=p_target_count
          AND reconciliation.source_sha256=p_source_sha256
          AND reconciliation.target_sha256=p_target_sha256
          AND reconciliation.lifecycle_state='ACCEPTED'
          AND reconciliation.requested_by=p_requested_by
          AND reconciliation.approved_by=p_approved_by
          AND reconciliation.created_at=p_accepted_at
          AND reconciliation.decided_at=p_accepted_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 reverse reconciliation replay differs';
    END IF;
    INSERT INTO documents_v1_reverse_receipts (
        reconciliation_id,tenant_id,source_generation,target_generation,
        source_count,target_count,source_sha256,target_sha256,
        requested_by,approved_by,accepted_at,receipt_sha256
    ) VALUES (
        p_reconciliation_id,p_tenant_id,p_source_generation,p_target_generation,
        p_source_count,p_target_count,p_source_sha256,p_target_sha256,
        p_requested_by,p_approved_by,p_accepted_at,v_receipt_sha
    ) ON CONFLICT (reconciliation_id) DO NOTHING;
    IF NOT EXISTS (
        SELECT 1 FROM documents_v1_reverse_receipts AS receipt
        WHERE receipt.reconciliation_id=p_reconciliation_id
          AND receipt.tenant_id=p_tenant_id
          AND receipt.source_generation=p_source_generation
          AND receipt.target_generation=p_target_generation
          AND receipt.source_count=p_source_count AND receipt.target_count=p_target_count
          AND receipt.source_sha256=p_source_sha256
          AND receipt.target_sha256=p_target_sha256
          AND receipt.requested_by=p_requested_by
          AND receipt.approved_by=p_approved_by
          AND receipt.accepted_at=p_accepted_at
          AND receipt.receipt_sha256=v_receipt_sha
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Documents V1 reverse receipt replay differs';
    END IF;
    RETURN v_receipt_sha;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION documents_v1_reverse_accept_v1(
    uuid,uuid,bigint,bigint,bigint,bigint,bytea,bytea,uuid,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_v1_reverse_accept_v1(
    uuid,uuid,bigint,bigint,bigint,bigint,bytea,bytea,uuid,uuid,timestamptz
) TO forgecrud_cutover_finalizer;
