INSERT INTO iam_lifecycle_delivery_worker_claims (
    artifact_id, artifact_ref, effect_id, effect_idempotency_key,
    message_id, state, claimed_at, outcome_unknown_at
)
VALUES ($1, $2, $3, $4, $5, 'claimed', $6, NULL)
ON CONFLICT (artifact_id) DO NOTHING
