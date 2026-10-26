SELECT operation_id,state,kit_claim_state
FROM platform.platformops_request_recovery_control_v1(
    $1,$2,$3,$4,$5,$6,$7,$8,$9
)
