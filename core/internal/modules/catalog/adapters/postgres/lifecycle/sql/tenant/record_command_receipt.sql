WITH inserted AS (
    INSERT INTO catalog_command_receipts (
        receipt_id, operation_id, package_id, component_id, owner_module,
        owner_generation, owner_contract, content_sha256, action,
        idempotency_key, completed_at
    ) VALUES (
        $1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::text,
        $6::bigint, $7::text, $8::text, $9::text, $10::text, $11::timestamptz
    ) ON CONFLICT DO NOTHING RETURNING receipt_id
)
SELECT receipt_id::text FROM inserted
UNION ALL
SELECT receipt_id::text FROM catalog_command_receipts WHERE receipt_id = $1::uuid
LIMIT 1
