SELECT request_sha256, item_id::text, result_revision,
       result_label, result_sort_order, result_visible, completed_at
FROM catalog_menu_override_receipts
WHERE actor_id = $1::uuid AND idempotency_key = $2::text
