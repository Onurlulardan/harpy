SELECT iam_notifications_session_selector_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1::uuid, $2::uuid, $3::timestamptz
)
