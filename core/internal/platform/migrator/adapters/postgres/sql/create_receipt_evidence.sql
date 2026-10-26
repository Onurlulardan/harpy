INSERT INTO platform.bootstrap_receipts (
    kind,
    run_id,
    environment_id,
    release_id,
    database_origin,
    database_system_identifier,
    database_oid,
    operation_digest,
    canonical_document,
    canonical_sha256,
    signature_key_id,
    issued_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12
)
ON CONFLICT (kind, run_id) DO NOTHING;
