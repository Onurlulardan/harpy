SELECT operation_id, restore_receipt_id, license_binding_id,
       previous_sequence, previous_sha256, accepted_sequence,
       accepted_sha256, authority_revision, replay
FROM platform.billing_commit_restored_license_v1(
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26, $27, $28
)
