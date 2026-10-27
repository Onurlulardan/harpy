SELECT identifier_id, user_id, user_revision, token_version,
       authorization_version, access_mode_id, access_mode_revision,
       portal_binding_id, portal_binding_revision, credential_version,
       credential_revision, algorithm, salt, tag, memory_kib, iterations,
       parallelism, legacy_bcrypt_hash, credential_created_at
FROM iam_login_credential_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1, $2, $3, $4, $5, $6
);
