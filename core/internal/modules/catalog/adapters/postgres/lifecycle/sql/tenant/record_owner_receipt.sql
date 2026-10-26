WITH inserted AS (
    INSERT INTO catalog_operation_steps (
        receipt_id, operation_id, component_id, owner_module, owner_generation,
        owner_contract, content_sha256, action, ready, completed_at, receipt_sha256
    ) VALUES (
        $1::uuid, $2::uuid, $3::uuid, $4::text, $5::bigint,
        $6::text, $7::text, $8::text, $9::boolean, $10::timestamptz, $11::text
    ) ON CONFLICT DO NOTHING
    RETURNING receipt_id, operation_id, component_id, owner_module, owner_generation,
              owner_contract, content_sha256, action, ready, completed_at
)
SELECT * FROM inserted
UNION ALL
SELECT receipt_id, operation_id, component_id, owner_module, owner_generation,
       owner_contract, content_sha256, action, ready, completed_at
FROM catalog_operation_steps WHERE receipt_id = $1::uuid
LIMIT 1
