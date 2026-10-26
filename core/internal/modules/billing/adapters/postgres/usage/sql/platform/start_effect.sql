UPDATE platform.billing_usage_reservations AS reservation
SET effect_started_at = $5,
    effect_receipt_id = $6,
    revision = reservation.revision + 1
WHERE reservation.authority_id = $1 AND reservation.tenant_id = $2
  AND reservation.id = $3 AND reservation.revision = $4
  AND reservation.state = 'reserved'
  AND reservation.effect_started_at IS NULL
  AND reservation.effect_receipt_id IS NULL
  AND $5 < reservation.expires_at
RETURNING reservation.id;
