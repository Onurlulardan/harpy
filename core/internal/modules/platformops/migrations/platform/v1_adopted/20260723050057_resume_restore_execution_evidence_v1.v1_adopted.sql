-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_claim_restore_execution_v5(
    p_operation_id uuid,p_attempt_id uuid,p_lease_token uuid,p_plan_sha256 text
) RETURNS TABLE(
    attempt_id uuid,lease_token uuid,attempt_generation bigint,
    restore_state text,operation_id uuid,recovery_set_id uuid,
    environment_id uuid,target_disposition text,plan_sha256 text,
    plan_bytes bytea,safety_recovery_set_id uuid,
    safety_recovery_set_sha256 text,waiver_reason text,probe_sha256 text,
    destructive_confirmed boolean,manifest_sha256 text,manifest_bytes bytea,
    previous_database_system_id uuid,component_inventory jsonb,
    control_database_locator_sha256 text,
    scratch_database_locator_sha256 text,drain_ack_sha256 text,
    switch_ack_sha256 text,archive_readback_sha256 text,
    new_database_system_id uuid,object_merkle_sha256 text,
    locator_remap_sha256 text,safety_evidence_sha256 text,
    database_now timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_claim_attempt_id uuid;
    v_claim_lease_token uuid;
    v_claim_attempt_generation bigint;
    v_claim_restore_state text;
    v_claim_operation_id uuid;
    v_claim_recovery_set_id uuid;
    v_claim_environment_id uuid;
    v_claim_target_disposition text;
    v_claim_plan_sha256 text;
    v_claim_plan_bytes bytea;
    v_claim_safety_recovery_set_id uuid;
    v_claim_safety_recovery_set_sha256 text;
    v_claim_waiver_reason text;
    v_claim_probe_sha256 text;
    v_claim_destructive_confirmed boolean;
    v_claim_manifest_sha256 text;
    v_claim_manifest_bytes bytea;
    v_claim_previous_database_system_id uuid;
    v_claim_component_inventory jsonb;
    v_claim_database_now timestamptz;
BEGIN
    SELECT claimed.attempt_id,claimed.lease_token,
           claimed.attempt_generation,claimed.restore_state,
           claimed.operation_id,claimed.recovery_set_id,
           claimed.environment_id,claimed.target_disposition,
           claimed.plan_sha256,claimed.plan_bytes,
           claimed.safety_recovery_set_id,
           claimed.safety_recovery_set_sha256,claimed.waiver_reason,
           claimed.probe_sha256,claimed.destructive_confirmed,
           claimed.manifest_sha256,claimed.manifest_bytes,
           claimed.previous_database_system_id,
           claimed.component_inventory,claimed.database_now
    INTO v_claim_attempt_id,v_claim_lease_token,
         v_claim_attempt_generation,v_claim_restore_state,
         v_claim_operation_id,v_claim_recovery_set_id,
         v_claim_environment_id,v_claim_target_disposition,
         v_claim_plan_sha256,v_claim_plan_bytes,
         v_claim_safety_recovery_set_id,
         v_claim_safety_recovery_set_sha256,v_claim_waiver_reason,
         v_claim_probe_sha256,v_claim_destructive_confirmed,
         v_claim_manifest_sha256,v_claim_manifest_bytes,
         v_claim_previous_database_system_id,
         v_claim_component_inventory,v_claim_database_now
    FROM platform.platformops_claim_restore_execution_v4(
        p_operation_id,p_attempt_id,p_lease_token,p_plan_sha256
    ) AS claimed;
    IF NOT FOUND THEN RETURN; END IF;
    UPDATE platform.platformops_restore_execution_attempts AS current_attempt
    SET drain_ack_sha256=prior.drain_ack_sha256,
        archive_readback_sha256=prior.archive_readback_sha256,
        restored_database_system_id=prior.restored_database_system_id,
        object_merkle_sha256=prior.object_merkle_sha256,
        locator_remap_sha256=prior.locator_remap_sha256,
        safety_evidence_sha256=prior.safety_evidence_sha256,
        control_database_locator_sha256=prior.control_database_locator_sha256,
        scratch_database_locator_sha256=prior.scratch_database_locator_sha256,
        switch_ack_sha256=prior.switch_ack_sha256
    FROM (
        SELECT previous_attempt.*
        FROM platform.platformops_restore_execution_attempts AS previous_attempt
        WHERE previous_attempt.operation_id=p_operation_id
          AND previous_attempt.generation<v_claim_attempt_generation
          AND previous_attempt.plan_sha256=p_plan_sha256
        ORDER BY previous_attempt.generation DESC
        LIMIT 1
    ) AS prior
    WHERE current_attempt.attempt_id=p_attempt_id
      AND current_attempt.operation_id=p_operation_id
      AND current_attempt.lease_token=p_lease_token
      AND current_attempt.lifecycle_state='active'
      AND current_attempt.drain_ack_sha256 IS NULL
      AND current_attempt.archive_readback_sha256 IS NULL
      AND current_attempt.restored_database_system_id IS NULL
      AND current_attempt.object_merkle_sha256 IS NULL
      AND current_attempt.locator_remap_sha256 IS NULL
      AND current_attempt.safety_evidence_sha256 IS NULL
      AND current_attempt.control_database_locator_sha256 IS NULL
      AND current_attempt.scratch_database_locator_sha256 IS NULL
      AND current_attempt.switch_ack_sha256 IS NULL;
    RETURN QUERY
    SELECT v_claim_attempt_id,v_claim_lease_token,
           v_claim_attempt_generation,v_claim_restore_state,
           v_claim_operation_id,v_claim_recovery_set_id,
           v_claim_environment_id,v_claim_target_disposition,
           v_claim_plan_sha256,v_claim_plan_bytes,
           v_claim_safety_recovery_set_id,
           v_claim_safety_recovery_set_sha256,v_claim_waiver_reason,
           v_claim_probe_sha256,v_claim_destructive_confirmed,
           v_claim_manifest_sha256,v_claim_manifest_bytes,
           v_claim_previous_database_system_id,v_claim_component_inventory,
           COALESCE(attempt.control_database_locator_sha256,''),
           COALESCE(attempt.scratch_database_locator_sha256,''),
           COALESCE(attempt.drain_ack_sha256,''),
           COALESCE(attempt.switch_ack_sha256,''),
           COALESCE(attempt.archive_readback_sha256,''),
           attempt.restored_database_system_id,
           COALESCE(attempt.object_merkle_sha256,''),
           COALESCE(attempt.locator_remap_sha256,''),
           COALESCE(attempt.safety_evidence_sha256,''),
           v_claim_database_now
    FROM platform.platformops_restore_execution_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id
      AND attempt.operation_id=p_operation_id
      AND attempt.lease_token=p_lease_token;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_claim_restore_execution_v5(
    uuid,uuid,uuid,text
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.platformops_claim_restore_execution_v4(
    uuid,uuid,uuid,text
) FROM harpy_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_claim_restore_execution_v5(
    uuid,uuid,uuid,text
) TO harpy_restore;
