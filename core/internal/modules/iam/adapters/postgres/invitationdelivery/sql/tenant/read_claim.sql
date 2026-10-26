SELECT artifact_ref, effect_id::text, effect_idempotency_key, message_id, state
FROM iam_invitation_delivery_worker_claims
WHERE artifact_ref = $1
FOR UPDATE
