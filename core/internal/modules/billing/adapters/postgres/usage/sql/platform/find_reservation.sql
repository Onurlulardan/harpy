SELECT reservation.id IS NOT NULL,
       COALESCE(reservation.id, '00000000-0000-0000-0000-000000000000'::uuid),
       COALESCE(reservation.decision_id, decision.id),
       COALESCE(reservation.entitlement_key, ''),
       COALESCE(reservation.meter_period_id, decision.meter_period_id),
       COALESCE(reservation.source_operation_id, decision.source_operation_id),
       COALESCE(reservation.amount, 0), COALESCE(reservation.state, ''),
       COALESCE(reservation.expires_at, decision.decided_at),
       reservation.effect_started_at, reservation.effect_receipt_id,
       reservation.resolution_receipt_id,
       COALESCE(commit_entry.amount_delta, 0),
       COALESCE(reservation.revision, 0),
       decision.id, decision.entitlement_key, decision.meter_period_id,
       decision.source_operation_id, decision.current_amount,
       decision.requested_amount, decision.limit_amount, decision.unlimited,
       decision.allowed, decision.soft_exceeded, decision.decided_at
FROM platform.billing_entitlement_decisions AS decision
LEFT JOIN platform.billing_usage_reservations AS reservation
  ON reservation.decision_id = decision.id
 AND reservation.authority_id = decision.authority_id
 AND reservation.tenant_id = decision.tenant_id
LEFT JOIN LATERAL (
    SELECT entry.amount_delta
    FROM platform.billing_usage_entries AS entry
    WHERE entry.authority_id = reservation.authority_id
      AND entry.tenant_id = reservation.tenant_id
      AND entry.reservation_id = reservation.id
      AND entry.entry_kind = 'commit'
) AS commit_entry ON true
WHERE decision.authority_id = $1 AND decision.tenant_id = $2
  AND decision.entitlement_key = $3 AND decision.source_operation_id = $4;
