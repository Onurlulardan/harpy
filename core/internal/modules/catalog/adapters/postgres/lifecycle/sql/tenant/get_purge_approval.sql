SELECT receipt_id,tenant_id,installation_id,package_id,operation_id,requested_by,
       approved_by,backup_receipt_id,backup_sha256,confirmation_slug,reason,
       purge_idempotency_key,request_sha256,approval_idempotency_key,
       step_up_authenticated_at,approved_at,expires_at,state,approval_sha256,
       consumed_operation_id,consumed_authority_sha256,consumed_at
FROM catalog_purge_approval_receipts_v2
WHERE receipt_id=$1::uuid
