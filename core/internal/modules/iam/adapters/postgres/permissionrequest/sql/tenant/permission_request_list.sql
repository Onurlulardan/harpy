SELECT *
FROM iam_permission_request_list_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1, $2, $3, $4
)
