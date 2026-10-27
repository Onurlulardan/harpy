WITH inserted AS (
    INSERT INTO platform.idempotency_receipts (
        receipt_id, record_kind, tenant_id, scope_kind, scope_id,
        authority_group_id, authority_group_generation,
        authority_domain, cutover_generation, writer_origin,
        request_sha256, retention_class,
        effect_id, effect_kind, effect_revision, effect_key,
        provider, operation, correlation_id, causation_id,
        effect_status, effect_error_class, effect_code,
        provider_reference, effect_result_ref, effect_result_sha256,
        completed_at, retain_until
    ) VALUES (
        $1, 'effect', $2, $3, $4,
        $5, $6, $7, $8, 'v2',
        $9, 'external_effect',
        $10, $11, $12, $13,
        $14, $15, $16, $17,
        $18, $19, $20,
        $21, $22, $23,
        $24::timestamptz, $24::timestamptz + interval '90 days'
    )
    ON CONFLICT DO NOTHING
    RETURNING receipt_id
)
SELECT EXISTS (SELECT 1 FROM inserted) OR EXISTS (
    SELECT 1
    FROM platform.idempotency_receipts AS receipt
    WHERE receipt.receipt_id = $1
      AND receipt.record_kind = 'effect'
      AND receipt.tenant_id IS NOT DISTINCT FROM $2::uuid
      AND receipt.scope_kind = $3
      AND receipt.scope_id = $4
      AND receipt.authority_group_id = $5
      AND receipt.authority_group_generation = $6
      AND receipt.authority_domain = $7
      AND receipt.cutover_generation = $8
      AND receipt.writer_origin = 'v2'
      AND receipt.request_sha256 = $9
      AND receipt.retention_class = 'external_effect'
      AND receipt.effect_id = $10
      AND receipt.effect_kind = $11
      AND receipt.effect_revision = $12
      AND receipt.effect_key = $13
      AND receipt.provider = $14
      AND receipt.operation = $15
      AND receipt.correlation_id = $16
      AND receipt.causation_id = $17
      AND receipt.effect_status = $18
      AND receipt.effect_error_class IS NOT DISTINCT FROM $19::text
      AND receipt.effect_code = $20
      AND receipt.provider_reference IS NOT DISTINCT FROM $21::text
      AND receipt.effect_result_ref IS NOT DISTINCT FROM $22::text
      AND receipt.effect_result_sha256 IS NOT DISTINCT FROM $23::text
      AND receipt.completed_at = $24::timestamptz
);
