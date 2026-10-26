SELECT EXISTS (
    SELECT 1
    FROM platform.idempotency_receipts AS receipt
    WHERE receipt.record_kind = 'effect'
      AND receipt.scope_kind = $1
      AND receipt.scope_id = $2
      AND receipt.authority_group_id = $3
      AND receipt.authority_group_generation = $4
      AND receipt.authority_domain = $5
      AND receipt.cutover_generation = $6
      AND receipt.writer_origin = 'v2'
      AND receipt.effect_key = $7
      AND receipt.effect_id = $8
      AND receipt.request_sha256 = $9
);
