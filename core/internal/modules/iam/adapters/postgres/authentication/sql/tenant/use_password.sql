SELECT user_revision, token_version, authorization_version,
       credential_version, credential_revision
FROM iam_password_use_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
);
