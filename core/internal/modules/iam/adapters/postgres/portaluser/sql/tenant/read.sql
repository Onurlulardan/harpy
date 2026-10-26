SELECT * FROM iam_portal_user_read_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1::uuid
);
