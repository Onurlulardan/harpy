SELECT *
FROM iam_invitation_read_v2(
    (pg_catalog.current_schemas(false))[2]::name, $1, $2
)
