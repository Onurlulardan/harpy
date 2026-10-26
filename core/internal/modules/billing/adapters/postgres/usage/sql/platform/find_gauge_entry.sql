SELECT entry.id, entry.decision_id, entry.entitlement_key,
       entry.meter_period_id, entry.owner_fact_id, entry.amount_delta,
       entry.gauge_after - entry.amount_delta, entry.gauge_after,
       entry.created_at, decision.id, decision.entitlement_key,
       decision.meter_period_id, decision.source_operation_id,
       decision.current_amount, decision.requested_amount,
       decision.limit_amount, decision.unlimited,
       decision.allowed, decision.soft_exceeded, decision.decided_at
FROM platform.billing_usage_entries AS entry
JOIN platform.billing_entitlement_decisions AS decision
  ON decision.id = entry.decision_id
 AND decision.authority_id = entry.authority_id
 AND decision.tenant_id = entry.tenant_id
WHERE entry.authority_id = $1 AND entry.tenant_id = $2
  AND entry.entitlement_key = $3 AND entry.owner_fact_id = $4
  AND entry.entry_kind = 'gauge_delta';
