SELECT *
FROM iam_permission_request_action_catalog_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1, $2, $3, $4
)
