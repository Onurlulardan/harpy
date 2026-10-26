WITH updated AS (
    UPDATE platform.billing_usage_totals AS total
    SET gauge_total = $9,
        revision = total.revision + 1,
        updated_at = GREATEST($10, total.updated_at + interval '1 microsecond')
    WHERE total.authority_id = $1 AND total.tenant_id = $2
      AND total.meter_period_id = $6 AND total.entitlement_key = $5
      AND total.meter_kind = 'gauge' AND total.gauge_total = $8
    RETURNING total.committed_total, total.reserved_total, total.gauge_total
), inserted AS (
    INSERT INTO platform.billing_usage_entries (
        id, authority_id, tenant_id, decision_id, entitlement_key,
        meter_period_id, reservation_id, entry_kind, amount_delta,
        committed_after, reserved_after, gauge_after, source_operation_id,
        owner_fact_id, evidence_sha256, occurred_at, created_at
    )
    SELECT $3, $1, $2, $4, $5, $6, NULL, 'gauge_delta', $7,
           updated.committed_total, updated.reserved_total, updated.gauge_total,
           NULL, $11, NULL, $10, $10
    FROM updated
    RETURNING id
)
SELECT id FROM inserted;
