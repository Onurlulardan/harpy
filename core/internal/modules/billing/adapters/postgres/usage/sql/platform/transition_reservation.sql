WITH changed AS (
    UPDATE platform.billing_usage_reservations AS reservation
    SET state = $7::text,
        resolution_receipt_id = CASE
          WHEN $7::text = 'reconcile_required' THEN NULL
          ELSE $9::uuid
        END,
        resolved_at = CASE
          WHEN $7::text = 'reconcile_required' THEN NULL
          ELSE $8::timestamptz
        END,
        revision = reservation.revision + 1
    WHERE reservation.authority_id = $1 AND reservation.tenant_id = $2
      AND reservation.id = $4 AND reservation.revision = $5
      AND reservation.state = $6
      AND $7::text IN ('committed', 'released', 'reconcile_required')
      AND (
          ($7::text = 'committed' AND $10::bigint > 0
            AND $10::bigint <= reservation.amount
            AND (($10::bigint = reservation.amount
                  AND $11::uuid = '00000000-0000-0000-0000-000000000000'::uuid)
              OR ($10::bigint < reservation.amount
                  AND $11::uuid <> '00000000-0000-0000-0000-000000000000'::uuid
                  AND $11::uuid <> $3::uuid)))
          OR ($7::text IN ('released', 'reconcile_required')
              AND $10::bigint = 0
              AND $11::uuid = '00000000-0000-0000-0000-000000000000'::uuid)
      )
    RETURNING reservation.id, reservation.decision_id,
              reservation.entitlement_key, reservation.meter_period_id,
              reservation.source_operation_id, reservation.amount,
              reservation.state
), updated AS (
    UPDATE platform.billing_usage_totals AS total
    SET committed_total = total.committed_total + CASE
          WHEN changed.state = 'committed' AND total.meter_kind = 'counter'
            THEN $10::bigint ELSE 0
        END,
        reserved_total = total.reserved_total - changed.amount,
        gauge_total = total.gauge_total + CASE
          WHEN changed.state = 'committed' AND total.meter_kind = 'gauge'
            THEN $10::bigint ELSE 0
        END,
        revision = total.revision + 1,
        updated_at = GREATEST(
          $8::timestamptz, total.updated_at + interval '1 microsecond'
        )
    FROM changed
    WHERE changed.state IN ('committed', 'released')
      AND total.authority_id = $1 AND total.tenant_id = $2
      AND total.meter_period_id = changed.meter_period_id
      AND total.entitlement_key = changed.entitlement_key
      AND total.meter_kind IN ('counter', 'gauge')
      AND total.reserved_total >= changed.amount
    RETURNING total.committed_total, total.reserved_total, total.gauge_total
), entry AS (
    INSERT INTO platform.billing_usage_entries (
        id, authority_id, tenant_id, decision_id, entitlement_key,
        meter_period_id, reservation_id, entry_kind, amount_delta,
        committed_after, reserved_after, gauge_after, source_operation_id,
        owner_fact_id, evidence_sha256, occurred_at, created_at
    )
    SELECT $3, $1, $2, changed.decision_id, changed.entitlement_key,
           changed.meter_period_id, changed.id,
           CASE WHEN changed.state = 'committed' THEN 'commit' ELSE 'release' END,
           CASE WHEN changed.state = 'committed'
             THEN $10::bigint ELSE -changed.amount END,
           updated.committed_total, updated.reserved_total, updated.gauge_total,
           changed.source_operation_id, NULL, NULL,
           $8::timestamptz, $8::timestamptz
    FROM changed CROSS JOIN updated
    RETURNING reservation_id
), remainder_entry AS (
    INSERT INTO platform.billing_usage_entries (
        id, authority_id, tenant_id, decision_id, entitlement_key,
        meter_period_id, reservation_id, entry_kind, amount_delta,
        committed_after, reserved_after, gauge_after, source_operation_id,
        owner_fact_id, evidence_sha256, occurred_at, created_at
    )
    SELECT $11::uuid, $1, $2, changed.decision_id, changed.entitlement_key,
           changed.meter_period_id, changed.id, 'release',
           -(changed.amount - $10::bigint), updated.committed_total,
           updated.reserved_total, updated.gauge_total,
           changed.source_operation_id, NULL, NULL,
           $8::timestamptz, $8::timestamptz
    FROM changed CROSS JOIN updated
    WHERE changed.state = 'committed' AND $10::bigint < changed.amount
    RETURNING reservation_id
)
SELECT changed.id
FROM changed
LEFT JOIN entry ON entry.reservation_id = changed.id
LEFT JOIN remainder_entry ON remainder_entry.reservation_id = changed.id
WHERE changed.state = 'reconcile_required'
   OR (entry.reservation_id IS NOT NULL
       AND ($10::bigint = changed.amount
            OR changed.state = 'released'
            OR remainder_entry.reservation_id IS NOT NULL));
