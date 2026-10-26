SELECT receipt_id, tenant_id, package_id, catalog_operation_id,
       cutover_generations, owner_manifest_root, database_snapshot_root,
       object_inventory_root, key_revision, completed_at,
       restore_verified_at, expires_at, verifier_version, legal_hold
FROM platform.backup_receipts
WHERE receipt_id = $1::uuid
