SELECT tenant_id, package_id, package_slug, operation_id, actor_id,
       requested_by, approved_by, backup_receipt_id, backup_state,
       approval_receipt_id,approval_sha256,backup_sha256, confirmation_slug,
       reason, idempotency_key, reference_epoch, step_up_authenticated_at,
       authorization_token_version,authorization_version,authorization_clock,
       authorization_permission_id,authorization_evidence_sha256,
       step_up_purpose_sha256,
       approval_step_up_at,approved_at,approval_expires_at,requested_at,
       cooling_until, authority_sha256
FROM catalog_purge_authorities_v2
WHERE operation_id = $1::uuid
