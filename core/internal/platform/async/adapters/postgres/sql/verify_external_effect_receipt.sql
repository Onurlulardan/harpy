SELECT EXISTS (
    SELECT 1
    FROM platform.idempotency_receipts AS receipt
    WHERE receipt.receipt_id = $7::uuid AND receipt.record_kind = 'effect'
      AND receipt.tenant_id = $8::uuid
      AND receipt.scope_kind = $1::text AND receipt.scope_id = $2::uuid
      AND receipt.authority_group_id = $3::text
      AND receipt.authority_group_generation = $4::bigint
      AND receipt.authority_domain = $5::text
      AND receipt.cutover_generation = $6::bigint
      AND receipt.writer_origin = 'v2'
      AND receipt.effect_id = $9::uuid AND receipt.effect_kind = $10::text
      AND receipt.effect_revision = $11::bigint
      AND receipt.effect_key = $12::text AND receipt.provider = $13::text
      AND receipt.operation = $14::text AND receipt.correlation_id = $15::text
      AND receipt.causation_id = $16::uuid AND receipt.effect_status = $17::text
      AND receipt.effect_error_class IS NOT DISTINCT FROM $18::text
      AND receipt.effect_code = $19::text
      AND receipt.provider_reference IS NOT DISTINCT FROM $20::text
      AND receipt.effect_result_ref IS NOT DISTINCT FROM $21::text
      AND receipt.effect_result_sha256 IS NOT DISTINCT FROM $22::text
      AND receipt.completed_at = $23::timestamptz
);
