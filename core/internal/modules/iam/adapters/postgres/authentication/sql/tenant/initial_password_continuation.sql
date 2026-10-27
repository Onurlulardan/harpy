SELECT selector, tenant_id, user_id, installation_operation_id,
       credential_id, credential_version, credential_revision,
       secret_key_id, secret_hmac, csrf_key_id, csrf_hmac,
       state, policy_attempts, revision, issued_at, expires_at,
       updated_at, consumed_at
FROM iam_initial_password_continuation_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,$2
)

