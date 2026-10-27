WITH inserted AS (
    INSERT INTO platform.billing_usage_reservations (
        id, authority_id, tenant_id, decision_id, entitlement_key,
        meter_period_id, source_operation_id, amount, state, expires_at,
        effect_started_at, effect_receipt_id, resolution_receipt_id,
        created_at, resolved_at, revision
    ) VALUES (
        $3, $1, $2, $4, $5, $6, $7, $8, 'reserved', $9,
        NULL, NULL, NULL, $10, NULL, 1
    )
    RETURNING id, decision_id, entitlement_key, meter_period_id,
              source_operation_id, amount
), updated AS (
    UPDATE platform.billing_usage_totals AS total
    SET reserved_total = total.reserved_total + inserted.amount,
        revision = total.revision + 1,
        updated_at = GREATEST($10, total.updated_at + interval '1 microsecond')
    FROM inserted
    WHERE total.authority_id = $1 AND total.tenant_id = $2
      AND total.meter_period_id = inserted.meter_period_id
      AND total.entitlement_key = inserted.entitlement_key
      AND total.meter_kind IN ('counter', 'gauge')
    RETURNING total.committed_total, total.reserved_total, total.gauge_total
), entry AS (
    INSERT INTO platform.billing_usage_entries (
        id, authority_id, tenant_id, decision_id, entitlement_key,
        meter_period_id, reservation_id, entry_kind, amount_delta,
        committed_after, reserved_after, gauge_after, source_operation_id,
        owner_fact_id, evidence_sha256, occurred_at, created_at
    )
    SELECT inserted.decision_id, $1, $2, inserted.decision_id,
           inserted.entitlement_key, inserted.meter_period_id, inserted.id,
           'reserve', inserted.amount, updated.committed_total,
           updated.reserved_total, updated.gauge_total,
           inserted.source_operation_id, NULL, NULL, $10, $10
    FROM inserted CROSS JOIN updated
    RETURNING reservation_id
)
SELECT reservation_id FROM entry;
