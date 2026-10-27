-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_recovery_seal_hold_v1(
    p_recovery_set_id uuid,
    p_inventory_sha256 bytea,
    p_inventory_count bigint
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_backup'
       OR pg_catalog.octet_length(p_inventory_sha256) <> 32
       OR p_inventory_count < 0 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'documents recovery seal denied';
    END IF;
    UPDATE documents_dr_holds AS hold
    SET lifecycle_state = 'VERIFIED',
        inventory_sha256 = p_inventory_sha256,
        inventory_count = p_inventory_count,
        updated_at = v_now
    WHERE hold.recovery_set_id = p_recovery_set_id
      AND hold.lifecycle_state IN ('HOLDING','COPYING')
      AND (
          SELECT pg_catalog.count(*)
          FROM documents_dr_copy_receipts AS receipt
          WHERE receipt.recovery_set_id = p_recovery_set_id
      ) = p_inventory_count;
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION documents_recovery_release_hold_v1(
    p_recovery_set_id uuid,
    p_inventory_sha256 bytea
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_backup' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'documents recovery release denied';
    END IF;
    UPDATE documents_dr_holds
    SET lifecycle_state = 'RELEASED', released_at = v_now, updated_at = v_now
    WHERE recovery_set_id = p_recovery_set_id
      AND lifecycle_state = 'VERIFIED'
      AND inventory_sha256 = p_inventory_sha256;
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION documents_recovery_fail_hold_v1(
    p_recovery_set_id uuid,
    p_failure_code text
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_backup'
       OR p_failure_code !~ '^[a-z0-9_]{1,64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'documents recovery failure denied';
    END IF;
    UPDATE documents_dr_holds
    SET lifecycle_state = 'FAILED', failure_code = p_failure_code,
        updated_at = pg_catalog.statement_timestamp()
    WHERE recovery_set_id = p_recovery_set_id
      AND lifecycle_state IN ('HOLDING','COPYING','VERIFIED');
    RETURN FOUND;
END;
$function$;
-- +goose StatementEnd
