SELECT * FROM iam_password_change_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11
)
