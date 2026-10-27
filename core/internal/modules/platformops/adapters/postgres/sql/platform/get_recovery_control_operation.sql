SELECT operation_id,environment_id,operation_kind,state,kit_claim_state,
       artifact_sha256,expires_at,failure_code,created_at
FROM platform.platformops_get_recovery_control_operation_v1($1)
