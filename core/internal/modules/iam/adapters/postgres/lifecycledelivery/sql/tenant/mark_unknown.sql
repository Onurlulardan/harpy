UPDATE iam_lifecycle_delivery_worker_claims
SET state = 'outcome_unknown', outcome_unknown_at = $4
WHERE artifact_ref = $1 AND effect_id = $2
  AND effect_idempotency_key = $3
  AND (state = 'claimed' OR
       (state = 'outcome_unknown' AND outcome_unknown_at = $4))
