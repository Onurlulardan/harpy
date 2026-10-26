SELECT operation_id, lease_token, license_id,
       licensee_organization_id, environment_id, delivery_kind,
       nonce, head_found, head_sequence, head_sha256,
       database_now, lease_expires_at
FROM platform.billing_license_refresh_claim_v1(
    $1, $2, $3, $4, $5, $6, $7
)
