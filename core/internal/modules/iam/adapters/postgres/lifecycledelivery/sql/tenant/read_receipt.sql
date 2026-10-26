SELECT receipt_id, artifact_ref, effect_idempotency_key, message_id,
       outcome, code, provider_reference, result_sha256, completed_at
FROM iam_lifecycle_delivery_worker_receipts
WHERE artifact_ref = $1
