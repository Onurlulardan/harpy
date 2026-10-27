SELECT receipt_id, artifact_ref, effect_idempotency_key, message_id,
       outcome, code, provider_reference, result_sha256, completed_at
FROM iam_invitation_delivery_worker_finalize_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
