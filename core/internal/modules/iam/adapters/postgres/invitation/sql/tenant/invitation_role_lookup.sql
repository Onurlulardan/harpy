SELECT *
FROM iam_invitation_role_lookup_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1, $2
)
