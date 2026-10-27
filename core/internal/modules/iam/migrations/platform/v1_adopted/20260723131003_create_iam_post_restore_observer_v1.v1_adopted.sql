-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_post_restore_observe_v1(
    p_environment_id uuid,p_restore_operation_id uuid
)
RETURNS TABLE(
    state text,receipt_sha256 text,reactivated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_reconciliation platform.iam_post_restore_reconciliations%ROWTYPE;
    v_admission_environment_id uuid;
    v_admission_restore_operation_id uuid;
    v_admission_state text;
    v_receipt_sha256 bytea;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_reactivation_observer'
       OR pg_catalog.current_setting('application_name') IS DISTINCT FROM
          'harpy-v2:core-operator-post-restore-reactivation-observer'
       OR p_environment_id IS NULL OR p_restore_operation_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='post-restore reactivation observation denied';
    END IF;
    SELECT * INTO v_reconciliation
    FROM platform.iam_post_restore_reconciliations
    WHERE environment_id=p_environment_id
      AND restore_operation_id=p_restore_operation_id;
    IF NOT FOUND OR NOT EXISTS (
        SELECT 1
        FROM platform.platformops_database_lineages AS lineage
        JOIN platform.platformops_database_lineage_events AS event
          ON event.environment_id=lineage.environment_id
         AND event.database_system_id=lineage.current_database_system_id
         AND event.event_kind='restored'
        WHERE lineage.environment_id=p_environment_id
          AND lineage.current_database_system_id=
              v_reconciliation.new_database_system_id
          AND event.operation_id=p_restore_operation_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='post-restore reactivation identity differs';
    END IF;
    SELECT admission.environment_id,admission.restore_operation_id,
           admission.state
    INTO v_admission_environment_id,v_admission_restore_operation_id,
         v_admission_state
    FROM platform.iam_post_restore_admission_v1(true) AS admission;
    IF NOT FOUND OR v_admission_environment_id<>p_environment_id
       OR v_admission_restore_operation_id<>p_restore_operation_id
       OR v_admission_state<>v_reconciliation.state
       OR v_reconciliation.state='fenced' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='post-restore reactivation admission differs';
    END IF;
    IF v_reconciliation.state='identity_reactivation_required' THEN
        RETURN QUERY SELECT 'pending'::text,NULL::text,NULL::timestamptz;
        RETURN;
    END IF;
    SELECT receipt.receipt_sha256 INTO v_receipt_sha256
    FROM platform.iam_post_restore_reactivation_receipts AS receipt
    WHERE receipt.restore_operation_id=p_restore_operation_id
      AND receipt.environment_id=p_environment_id;
    IF NOT FOUND OR v_reconciliation.state<>'reactivated'
       OR v_reconciliation.reactivated_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='post-restore reactivation receipt differs';
    END IF;
    RETURN QUERY SELECT 'reactivated'::text,
        pg_catalog.encode(v_receipt_sha256,'hex'),
        v_reconciliation.reactivated_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_post_restore_observe_v1(
    uuid,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_post_restore_observe_v1(
    uuid,uuid
) TO harpy_reactivation_observer;
