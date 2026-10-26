SELECT iam_security_epoch_assert_v1(
    (pg_catalog.current_schemas(false))[2]::name, $1
);
