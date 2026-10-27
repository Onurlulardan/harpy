-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_request_restore_v1(
    p_operation_id uuid,
    p_recovery_set_id uuid,
    p_environment_id uuid,
    p_target_disposition text,
    p_plan_sha256 text,
    p_plan_bytes bytea,
    p_safety_recovery_set_id uuid,
    p_safety_recovery_set_sha256 text,
    p_waiver_reason text,
    p_probe_sha256 text,
    p_destructive_confirmed boolean,
    p_created_at timestamptz
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing platform.platformops_restore_operations%ROWTYPE;
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM platform.platformops_recovery_sets
         WHERE recovery_set_id = p_recovery_set_id
           AND environment_id = p_environment_id
           AND state = 'protected'
    ) THEN
        RAISE EXCEPTION 'protected same-environment recovery set is missing'
            USING ERRCODE = 'P0002';
    END IF;
    IF p_target_disposition = 'existing_readable' AND NOT EXISTS (
        SELECT 1 FROM platform.platformops_recovery_sets
         WHERE recovery_set_id = p_safety_recovery_set_id
           AND recovery_set_id <> p_recovery_set_id
           AND environment_id = p_environment_id
           AND state = 'protected'
           AND manifest_sha256 = p_safety_recovery_set_sha256
    ) THEN
        RAISE EXCEPTION 'readable restore requires protected safety set'
            USING ERRCODE = '55000';
    END IF;
    SELECT * INTO v_existing FROM platform.platformops_restore_operations
     WHERE operation_id = p_operation_id;
    IF FOUND THEN
        IF v_existing.recovery_set_id <> p_recovery_set_id
           OR v_existing.environment_id <> p_environment_id
           OR v_existing.target_disposition <> p_target_disposition
           OR v_existing.plan_sha256 <> p_plan_sha256
           OR v_existing.plan_bytes <> p_plan_bytes
           OR v_existing.safety_recovery_set_id IS DISTINCT FROM p_safety_recovery_set_id
           OR v_existing.safety_recovery_set_sha256 IS DISTINCT FROM p_safety_recovery_set_sha256
           OR v_existing.waiver_reason IS DISTINCT FROM p_waiver_reason
           OR v_existing.probe_sha256 IS DISTINCT FROM p_probe_sha256
           OR v_existing.destructive_confirmed <> p_destructive_confirmed
           OR v_existing.created_at <> p_created_at THEN
            RAISE EXCEPTION 'restore request conflicts' USING ERRCODE = '23505';
        END IF;
        RETURN;
    END IF;
    INSERT INTO platform.platformops_restore_operations(
        operation_id, recovery_set_id, environment_id, target_disposition,
        state, plan_sha256, plan_bytes, safety_recovery_set_id,
        safety_recovery_set_sha256, waiver_reason, probe_sha256,
        destructive_confirmed, created_at, updated_at
    ) VALUES (
        p_operation_id, p_recovery_set_id, p_environment_id, p_target_disposition,
        'plan_ready', p_plan_sha256, p_plan_bytes, p_safety_recovery_set_id,
        p_safety_recovery_set_sha256, p_waiver_reason, p_probe_sha256,
        p_destructive_confirmed, p_created_at, p_created_at
    );
END;
$function$;

CREATE FUNCTION platform.platformops_advance_restore_v1(
    p_operation_id uuid,
    p_expected_state text,
    p_next_state text,
    p_updated_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    UPDATE platform.platformops_restore_operations
       SET state = p_next_state, updated_at = p_updated_at
     WHERE operation_id = p_operation_id AND state = p_expected_state
       AND p_next_state IN ('draining','restoring','verifying','switched','failed','fenced')
    ;
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION platform.platformops_complete_restore_v1(
    p_operation_id uuid,
    p_receipt_sha256 text,
    p_receipt_bytes bytea,
    p_new_database_system_id uuid,
    p_completed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    UPDATE platform.platformops_restore_operations
       SET state = 'completed', receipt_sha256 = p_receipt_sha256,
           receipt_bytes = p_receipt_bytes,
           new_database_system_id = p_new_database_system_id,
           completed_at = p_completed_at, updated_at = p_completed_at
     WHERE operation_id = p_operation_id AND state = 'switched'
       AND p_receipt_sha256 ~ '^[0-9a-f]{64}$'
       AND octet_length(p_receipt_bytes) BETWEEN 1 AND 2097152
    ;
    RETURN FOUND;
END;
$function$;
-- +goose StatementEnd
