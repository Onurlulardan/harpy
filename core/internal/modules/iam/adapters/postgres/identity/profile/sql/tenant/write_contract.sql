SELECT iam_profile_write_contract_read_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::text
);
