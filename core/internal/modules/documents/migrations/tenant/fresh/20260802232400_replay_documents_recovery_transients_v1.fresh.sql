-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION documents_recovery_seal_hold_v2(
    p_recovery_set_id uuid,
    p_inventory_sha256 bytea,
    p_inventory_count bigint
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
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
        updated_at = pg_catalog.statement_timestamp()
    WHERE hold.recovery_set_id = p_recovery_set_id
      AND hold.lifecycle_state IN ('HOLDING','COPYING')
      AND (
          SELECT pg_catalog.count(*)
          FROM documents_dr_copy_receipts AS receipt
          WHERE receipt.recovery_set_id = p_recovery_set_id
      ) = p_inventory_count;
    IF FOUND THEN
        RETURN true;
    END IF;
    RETURN EXISTS (
        SELECT 1
        FROM documents_dr_holds AS hold
        WHERE hold.recovery_set_id = p_recovery_set_id
          AND hold.lifecycle_state IN ('VERIFIED','RELEASED')
          AND hold.inventory_sha256 = p_inventory_sha256
          AND hold.inventory_count = p_inventory_count
          AND (
              SELECT pg_catalog.count(*)
              FROM documents_dr_copy_receipts AS receipt
              WHERE receipt.recovery_set_id = p_recovery_set_id
          ) = p_inventory_count
    );
END;
$function$;
CREATE FUNCTION documents_recovery_release_hold_v2(
    p_recovery_set_id uuid,
    p_inventory_sha256 bytea
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_backup' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'documents recovery release denied';
    END IF;
    UPDATE documents_dr_holds
    SET lifecycle_state = 'RELEASED',
        released_at = pg_catalog.statement_timestamp(),
        updated_at = pg_catalog.statement_timestamp()
    WHERE recovery_set_id = p_recovery_set_id
      AND lifecycle_state = 'VERIFIED'
      AND inventory_sha256 = p_inventory_sha256;
    IF FOUND THEN
        RETURN true;
    END IF;
    RETURN EXISTS (
        SELECT 1 FROM documents_dr_holds
        WHERE recovery_set_id = p_recovery_set_id
          AND lifecycle_state = 'RELEASED'
          AND inventory_sha256 = p_inventory_sha256
    );
END;
$function$;
CREATE FUNCTION documents_recovery_fail_hold_v2(
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
    SET lifecycle_state = 'FAILED',
        failure_code = p_failure_code,
        updated_at = pg_catalog.statement_timestamp()
    WHERE recovery_set_id = p_recovery_set_id
      AND lifecycle_state IN ('HOLDING','COPYING','VERIFIED');
    IF FOUND THEN
        RETURN true;
    END IF;
    RETURN EXISTS (
        SELECT 1 FROM documents_dr_holds
        WHERE recovery_set_id = p_recovery_set_id
          AND lifecycle_state = 'FAILED'
          AND failure_code = p_failure_code
    );
END;
$function$;
CREATE FUNCTION documents_recovery_terminalize_transients_v2(
    p_restore_operation_id uuid
)
RETURNS TABLE(
    cancelled_uploads bigint,
    expired_archives bigint,
    cancelled_archives bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_restore'
       OR p_restore_operation_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'documents transient restore policy denied';
    END IF;
    WITH changed AS (
        UPDATE documents_uploads
        SET lifecycle_state = 'CANCELLED',
            lease_owner_id = NULL,
            lease_expires_at = NULL,
            cancel_requested_at = v_now,
            cleanup_required = false,
            updated_at = v_now
        WHERE lifecycle_state NOT IN ('READY','CANCELLED','EXPIRED')
        RETURNING id
    )
    INSERT INTO documents_restore_transient_dispositions(
        restore_operation_id,target_kind,target_id,disposition,disposed_at
    )
    SELECT p_restore_operation_id,'UPLOAD',id,'CANCELLED',v_now
    FROM changed
    ON CONFLICT DO NOTHING;
    WITH changed AS (
        UPDATE documents_archives
        SET lifecycle_state = 'EXPIRED',updated_at = v_now
        WHERE lifecycle_state = 'READY'
        RETURNING id
    )
    INSERT INTO documents_restore_transient_dispositions(
        restore_operation_id,target_kind,target_id,disposition,disposed_at
    )
    SELECT p_restore_operation_id,'ARCHIVE',id,'EXPIRED',v_now
    FROM changed
    ON CONFLICT DO NOTHING;

    WITH changed AS (
        UPDATE documents_archives
        SET lifecycle_state = 'CANCELLED',updated_at = v_now
        WHERE lifecycle_state = 'PENDING'
        RETURNING id
    )
    INSERT INTO documents_restore_transient_dispositions(
        restore_operation_id,target_kind,target_id,disposition,disposed_at
    )
    SELECT p_restore_operation_id,'ARCHIVE',id,'CANCELLED',v_now
    FROM changed
    ON CONFLICT DO NOTHING;

    SELECT count(*) FILTER (
               WHERE target_kind = 'UPLOAD' AND disposition = 'CANCELLED'
           ),
           count(*) FILTER (
               WHERE target_kind = 'ARCHIVE' AND disposition = 'EXPIRED'
           ),
           count(*) FILTER (
               WHERE target_kind = 'ARCHIVE' AND disposition = 'CANCELLED'
           )
    INTO cancelled_uploads,expired_archives,cancelled_archives
    FROM documents_restore_transient_dispositions
    WHERE restore_operation_id = p_restore_operation_id;
    RETURN NEXT;
END;
$function$;
REVOKE ALL ON FUNCTION documents_recovery_seal_hold_v2(uuid,bytea,bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_recovery_seal_hold_v2(uuid,bytea,bigint) TO harpy_backup_reader;
REVOKE ALL ON FUNCTION documents_recovery_release_hold_v2(uuid,bytea) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_recovery_release_hold_v2(uuid,bytea) TO harpy_backup_reader;
REVOKE ALL ON FUNCTION documents_recovery_fail_hold_v2(uuid,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_recovery_fail_hold_v2(uuid,text) TO harpy_backup_reader;
REVOKE ALL ON FUNCTION documents_recovery_terminalize_transients_v2(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION documents_recovery_terminalize_transients_v2(uuid) TO harpy_restore;
-- +goose StatementEnd
