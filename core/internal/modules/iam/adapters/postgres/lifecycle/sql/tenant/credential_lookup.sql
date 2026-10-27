SELECT * FROM iam_lifecycle_credential_lookup_v2(
    (pg_catalog.current_schemas(false))[2]::name, $1, $2, $3
)
