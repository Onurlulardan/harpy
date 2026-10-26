SELECT platform.iam_installation_security_epoch_initialize_v2(
    $1::uuid, $2::uuid, $3::uuid, $4::name,
    pg_catalog.statement_timestamp()
);
