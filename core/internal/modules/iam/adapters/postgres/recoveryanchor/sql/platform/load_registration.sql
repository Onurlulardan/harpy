SELECT operation_id,anchor_id,challenge,rp_id,allowed_origins,
       allowed_origins_sha256,expires_at,state
FROM platform.iam_recovery_anchor_registration_v1($1::uuid,$2::uuid,$3::uuid);
