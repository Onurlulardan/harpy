SELECT revision, issued_at, expires_at, credential_id, rp_id,
       user_verification_required
FROM iam_reauthentication_start_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18
);
