SELECT group_id, group_state, group_revision
FROM iam_permission_request_review_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1, $2, $3, $4, $5, $6
)
