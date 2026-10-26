SELECT artifact_id, invitation_id, token_id, artifact_ref, recipient,
       kms_key_id, nonce, ciphertext, authentication_tag, aad_sha256,
       ciphertext_sha256, expires_at
FROM iam_invitation_delivery_worker_claim_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1, $2, $3, $4, $5
)
