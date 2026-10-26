WITH inserted AS (
    INSERT INTO catalog_tombstones (
        package_id, slug, package_sha256, reference_epoch, owner_receipt_root, purged_at
    ) VALUES (
        $1::uuid, $2::text, $3::text, $4::bigint, $5::text, $6::timestamptz
    ) ON CONFLICT DO NOTHING
    RETURNING package_id, slug, package_sha256, reference_epoch, owner_receipt_root
)
SELECT package_id::text, slug, package_sha256, reference_epoch, owner_receipt_root
FROM inserted
UNION ALL
SELECT package_id::text, slug, package_sha256, reference_epoch, owner_receipt_root
FROM catalog_tombstones
WHERE package_id = $1::uuid
LIMIT 1
