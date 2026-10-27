SELECT user_id, access_mode_id, user_revision, token_version,
       authorization_version, access_mode_revision, credential_version,
       credential_revision, continuation_revision, installation_revision
FROM iam_initial_password_complete_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12
)

