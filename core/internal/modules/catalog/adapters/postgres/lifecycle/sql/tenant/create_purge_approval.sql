SELECT receipt_id,tenant_id,installation_id,package_id,operation_id,requested_by,
       approved_by,backup_receipt_id,backup_sha256,confirmation_slug,reason,
       purge_idempotency_key,request_sha256,approval_idempotency_key,
       step_up_authenticated_at,approved_at,expires_at,state,approval_sha256,
       consumed_operation_id,consumed_authority_sha256,consumed_at
FROM catalog_create_purge_approval_v2(
    $1::name,$2::uuid,$3::uuid,$4::uuid,$5::uuid,$6::uuid,$7::uuid,$8::uuid,
    $9::uuid,$10::text,$11::text,$12::text,$13::text,$14::text,$15::text,
    $16::timestamptz,$17::timestamptz,$18::timestamptz,$19::text
)
