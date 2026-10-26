SELECT operation_id,tenant_id,owner_user_id,anchor_id,
challenge,credential_id,public_key_cose,rp_id,sign_count,expires_at,replayed
FROM platform.iam_post_restore_challenge_v1($1,$2,$3,$4,$5,$6,$7,$8,$9)
