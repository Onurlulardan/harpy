WITH inserted AS (
    INSERT INTO catalog_menu_override_receipts (
        receipt_id, actor_id, idempotency_key, request_sha256, item_id,
        result_revision, result_label, result_sort_order, result_visible, completed_at
    ) VALUES (
        $1::uuid, $2::uuid, $3::text, $4::text, $5::uuid,
        $6::bigint, $7::text, $8::integer, $9::boolean, $10::timestamptz
    ) ON CONFLICT DO NOTHING
    RETURNING request_sha256, item_id, result_revision, result_label,
              result_sort_order, result_visible, completed_at
)
SELECT request_sha256, item_id::text, result_revision, result_label,
       result_sort_order, result_visible, completed_at, true AS inserted
FROM inserted
UNION ALL
SELECT request_sha256, item_id::text, result_revision, result_label,
       result_sort_order, result_visible, completed_at, false AS inserted
FROM catalog_menu_override_receipts
WHERE actor_id = $2::uuid AND idempotency_key = $3::text
LIMIT 1
