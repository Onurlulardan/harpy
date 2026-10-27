SELECT reservation.id, reservation.decision_id, reservation.entitlement_key,
       reservation.meter_period_id, reservation.source_operation_id,
       reservation.amount, reservation.state, reservation.expires_at,
       reservation.effect_started_at, reservation.effect_receipt_id,
       reservation.resolution_receipt_id,
       COALESCE(commit_entry.amount_delta, 0), reservation.revision
FROM platform.billing_usage_reservations AS reservation
LEFT JOIN LATERAL (
    SELECT entry.amount_delta
    FROM platform.billing_usage_entries AS entry
    WHERE entry.authority_id = reservation.authority_id
      AND entry.tenant_id = reservation.tenant_id
      AND entry.reservation_id = reservation.id
      AND entry.entry_kind = 'commit'
) AS commit_entry ON true
WHERE reservation.authority_id = $1 AND reservation.tenant_id = $2
  AND reservation.id = $3
FOR UPDATE OF reservation;
