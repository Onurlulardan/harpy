-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_post_restore_reconcile_v1(
    p_restore_operation_id uuid,p_environment_id uuid,p_recovery_set_id uuid,
    p_tenant_id uuid,p_owner_user_id uuid,p_anchor_id uuid,p_new_database_system_id uuid,
    p_security_epoch bytea,p_restore_receipt_sha256 bytea,p_protected_head_sha256 bytea,
    p_custody_generation bigint,p_custody_result_sha256 bytea,p_receipt jsonb,p_occurred_at timestamptz
)
RETURNS TABLE(state text,replayed boolean) LANGUAGE plpgsql SECURITY DEFINER
SET search_path TO pg_catalog AS $function$
DECLARE v_existing platform.iam_post_restore_reconciliations%ROWTYPE;
v_epoch bytea;
BEGIN
    IF p_restore_operation_id IS NULL OR p_environment_id IS NULL OR
       p_recovery_set_id IS NULL OR p_tenant_id IS NULL OR p_owner_user_id IS NULL OR
       p_anchor_id IS NULL OR p_new_database_system_id IS NULL OR
       pg_catalog.octet_length(p_security_epoch)<>32 OR
       pg_catalog.octet_length(p_restore_receipt_sha256)<>32 OR
       pg_catalog.octet_length(p_protected_head_sha256)<>32 OR
       p_custody_generation<1 OR pg_catalog.octet_length(p_custody_result_sha256)<>32 OR
       pg_catalog.jsonb_typeof(p_receipt)<>'object' THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='post-restore reconcile input is invalid';
    END IF;
    SELECT * INTO v_existing FROM platform.iam_post_restore_reconciliations
     WHERE restore_operation_id=p_restore_operation_id FOR UPDATE;
    IF FOUND THEN
        IF v_existing.environment_id<>p_environment_id OR
           v_existing.recovery_set_id<>p_recovery_set_id OR
           v_existing.tenant_id<>p_tenant_id OR v_existing.owner_user_id<>p_owner_user_id OR
           v_existing.anchor_id<>p_anchor_id OR
           v_existing.new_database_system_id<>p_new_database_system_id OR
           v_existing.security_epoch<>p_security_epoch OR
           v_existing.restore_receipt_sha256<>p_restore_receipt_sha256 OR
           v_existing.protected_head_sha256<>p_protected_head_sha256 OR
           v_existing.custody_generation<>p_custody_generation OR
           v_existing.custody_result_sha256<>p_custody_result_sha256 OR
           v_existing.receipt<>p_receipt THEN
            UPDATE platform.iam_post_restore_reconciliations SET state='fenced',updated_at=p_occurred_at
             WHERE restore_operation_id=p_restore_operation_id;
		    RETURN QUERY SELECT 'fenced'::text,false;
		    RETURN;
        END IF;
        RETURN QUERY SELECT v_existing.state,true;
        RETURN;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM platform.iam_recovery_anchors AS anchor
         WHERE anchor.anchor_id=p_anchor_id AND anchor.environment_id=p_environment_id
           AND anchor.tenant_id=p_tenant_id AND anchor.owner_user_id=p_owner_user_id
           AND anchor.state='active' AND anchor.user_verification='required'
           AND NOT anchor.backup_eligible AND NOT anchor.backup_state
           AND anchor.protected_state_sha256=p_protected_head_sha256
    ) OR NOT platform.iam_platform_current_owner_v1(p_tenant_id,p_owner_user_id) THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='post-restore protected owner binding is invalid';
    END IF;
    SELECT security_epoch INTO v_epoch FROM platform.iam_security_epochs
     WHERE environment_id=p_environment_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P0002',MESSAGE='IAM security epoch is missing';
    END IF;
    IF v_epoch<>p_security_epoch THEN
        UPDATE platform.iam_security_epochs SET security_epoch=p_security_epoch,
            generation=generation+1,rotated_at=p_occurred_at
         WHERE environment_id=p_environment_id;
    END IF;
    DELETE FROM platform.iam_refresh_locators WHERE tenant_id=p_tenant_id;
    INSERT INTO platform.iam_post_restore_reconciliations(
        restore_operation_id,environment_id,recovery_set_id,tenant_id,owner_user_id,anchor_id,
        new_database_system_id,security_epoch,restore_receipt_sha256,protected_head_sha256,
        custody_generation,custody_result_sha256,receipt,state,created_at,updated_at
    ) VALUES (p_restore_operation_id,p_environment_id,p_recovery_set_id,p_tenant_id,
        p_owner_user_id,p_anchor_id,p_new_database_system_id,p_security_epoch,
        p_restore_receipt_sha256,p_protected_head_sha256,p_custody_generation,
        p_custody_result_sha256,p_receipt,'identity_reactivation_required',p_occurred_at,p_occurred_at);
    RETURN QUERY SELECT 'identity_reactivation_required'::text,false;
END;
$function$;

CREATE FUNCTION platform.iam_post_restore_admission_v1(p_probe boolean)
RETURNS TABLE(state text,environment_id uuid,restore_operation_id uuid)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog AS $function$
BEGIN
IF p_probe IS DISTINCT FROM true THEN
    RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='post-restore admission probe is invalid';
END IF;
RETURN QUERY
WITH current_restore AS (
	SELECT event.operation_id,lineage.environment_id,lineage.current_database_system_id
	  FROM platform.platformops_database_lineages AS lineage
	  JOIN platform.platformops_database_lineage_events AS event
	    ON event.environment_id=lineage.environment_id
	   AND event.database_system_id=lineage.current_database_system_id
	   AND event.event_kind='restored'
	 ORDER BY event.generation DESC LIMIT 1
), restored_admission AS (
	SELECT COALESCE(reconciliation.state,'fenced') AS state,
	       restore.environment_id,restore.operation_id
	  FROM current_restore AS restore
	  LEFT JOIN platform.iam_post_restore_reconciliations AS reconciliation
	    ON reconciliation.restore_operation_id=restore.operation_id
	   AND reconciliation.environment_id=restore.environment_id
	   AND reconciliation.new_database_system_id=restore.current_database_system_id
)
SELECT admission.state,admission.environment_id,admission.operation_id
  FROM restored_admission AS admission
UNION ALL
SELECT reconciliation.state,reconciliation.environment_id,reconciliation.restore_operation_id
 FROM (SELECT candidate.* FROM platform.iam_post_restore_reconciliations AS candidate
		 ORDER BY candidate.created_at DESC LIMIT 1) AS reconciliation
 WHERE NOT EXISTS (SELECT 1 FROM current_restore)
;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_post_restore_reconcile_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,bytea,bytea,bytea,bigint,bytea,jsonb,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_post_restore_admission_v1(boolean) FROM PUBLIC;
