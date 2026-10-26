SELECT * FROM iam_password_reset_delivery_target_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1,$2
)
