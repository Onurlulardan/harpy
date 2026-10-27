SELECT claim.id, claim.subscription_id, claim.offering_version_id,
       claim.source_operation_id, claim.claimed_at
FROM platform.billing_trial_claims AS claim
WHERE claim.authority_id = $1 AND claim.tenant_id = $2
  AND claim.offering_version_id = $3;
