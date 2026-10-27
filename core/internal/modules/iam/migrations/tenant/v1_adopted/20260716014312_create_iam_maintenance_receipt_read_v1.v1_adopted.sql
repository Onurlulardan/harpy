-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_maintenance_receipt_read_v1(
    p_scope_schema name, p_tenant_id uuid, p_job_id uuid,
    p_kind text, p_payload_revision integer,
    p_payload_reference text, p_payload_sha256 bytea
)
RETURNS TABLE (
    successor_job_id uuid, successor_run_at timestamptz,
    successor_reference text, successor_sha256 bytea,
    successor_dedupe_key text, successor_priority integer
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_kind text;
    v_payload_revision integer;
    v_payload_reference text;
    v_payload_sha256 bytea;
    v_successor_job_id uuid;
    v_successor_run_at timestamptz;
    v_successor_reference text;
    v_successor_sha256 bytea;
    v_successor_dedupe_key text;
    v_successor_priority integer;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS NULL OR p_job_id IS NULL OR NOT EXISTS (
           SELECT 1 FROM platform.tenants AS tenant
           WHERE tenant.id = p_tenant_id AND tenant.schema_name = p_scope_schema
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM maintenance receipt authority is invalid';
    END IF;
    SELECT receipt.kind, receipt.payload_revision,
           receipt.payload_reference, receipt.payload_sha256,
           receipt.successor_job_id, receipt.successor_run_at,
           receipt.successor_payload_reference,
           receipt.successor_payload_sha256,
           receipt.successor_dedupe_key, receipt.successor_priority
    INTO v_kind, v_payload_revision, v_payload_reference, v_payload_sha256,
         v_successor_job_id, v_successor_run_at, v_successor_reference,
         v_successor_sha256, v_successor_dedupe_key, v_successor_priority
    FROM iam_maintenance_receipts AS receipt
    WHERE receipt.job_id = p_job_id;
    IF NOT FOUND THEN
        RETURN;
    END IF;
    IF v_kind IS DISTINCT FROM p_kind OR
       v_payload_revision IS DISTINCT FROM p_payload_revision OR
       v_payload_reference IS DISTINCT FROM p_payload_reference OR
       v_payload_sha256 IS DISTINCT FROM p_payload_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'IAM maintenance receipt identity conflicts';
    END IF;
    RETURN QUERY SELECT
        v_successor_job_id, v_successor_run_at, v_successor_reference,
        v_successor_sha256, v_successor_dedupe_key, v_successor_priority;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_maintenance_receipt_read_v1(
    name,uuid,uuid,text,integer,text,bytea
) FROM PUBLIC;
