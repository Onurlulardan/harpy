SELECT receipt_id::text, operation_id::text, component_id::text,
       owner_module, owner_generation, contract, content_sha256,
       action, ready, completed_at
FROM catalog_demo_component_apply_v1(
    $1::uuid,$2::uuid,$3::uuid,$4::text,$5::bigint,$6::text,
    $7::text,$8::text,$9::bigint,$10::text,$11::timestamptz
)
