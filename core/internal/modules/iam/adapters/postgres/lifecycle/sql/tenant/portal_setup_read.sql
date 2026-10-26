SELECT * FROM iam_portal_setup_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,$4,$5,$6
)
