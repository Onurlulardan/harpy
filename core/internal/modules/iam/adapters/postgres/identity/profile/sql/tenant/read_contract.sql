SELECT iam_profile_read_contract_v2(
    (pg_catalog.current_schemas(false))[2]::name,
    $1::text
);
