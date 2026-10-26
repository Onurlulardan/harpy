SELECT tenant_id,owner_user_id,anchor_id,challenge,
credential_id,public_key_cose,rp_id,sign_count,context_sha256
FROM platform.iam_post_restore_challenge_context_v1($1,$2,$3,$4,$5)
