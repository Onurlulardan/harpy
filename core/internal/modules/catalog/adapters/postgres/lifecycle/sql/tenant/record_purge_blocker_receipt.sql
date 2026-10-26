WITH inserted AS (
    INSERT INTO catalog_purge_blocker_receipts (
        receipt_id, operation_id, component_id, reference_epoch,
        catalog_version, zero_digest, completed_at
    ) VALUES ($1::uuid, $2::uuid, $3::uuid, $4::bigint, $5::text, $6::text, $7::timestamptz)
    ON CONFLICT DO NOTHING
    RETURNING receipt_id, operation_id, component_id, reference_epoch,
              catalog_version, zero_digest, completed_at
)
SELECT * FROM inserted
UNION ALL
SELECT receipt_id, operation_id, component_id, reference_epoch,
       catalog_version, zero_digest, completed_at
FROM catalog_purge_blocker_receipts WHERE receipt_id = $1::uuid
LIMIT 1
