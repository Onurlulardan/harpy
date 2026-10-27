SELECT identifier_id, user_id, user_revision, token_version,
       authorization_version, access_mode_id, access_mode_revision,
       installation_operation_id, credential_id, credential_version,
       credential_revision, algorithm, salt, tag, memory_kib, iterations,
       parallelism, credential_created_at
FROM iam_initial_password_credential_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5
)

