SELECT anchor_id,generation,credential_id,rp_id,user_verification,
       backup_eligible,backup_state,state
FROM platform.iam_recovery_anchor_list_v1($1::uuid,$2::uuid,$3::uuid);
