SELECT operation_id, offering_version_id, revision,
       offering_sha256, replay
FROM platform.billing_publish_license_derived_offering_v1(
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15
)
