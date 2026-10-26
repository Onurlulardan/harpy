INSERT INTO iam_lifecycle_delivery_worker_receipts (
    receipt_id, artifact_id, purpose, subject_user_id, resource_id, token_id,
    artifact_ref, effect_id, effect_idempotency_key, message_id, outcome, code,
    provider_reference, result_sha256, completed_at
)
SELECT $1, artifact.id, artifact.purpose, artifact.subject_user_id,
       artifact.resource_id, artifact.token_id, artifact.artifact_ref,
       $4, $5, $6, $7, $8, $9, $10, $11
FROM iam_lifecycle_delivery_artifacts AS artifact
JOIN iam_lifecycle_delivery_worker_claims AS claim
  ON claim.artifact_id = artifact.id
WHERE artifact.id = $2 AND artifact.artifact_ref = $3
  AND claim.effect_id = $4 AND claim.effect_idempotency_key = $5
  AND claim.state = 'claimed'
ON CONFLICT (artifact_id) DO NOTHING
