SELECT result.revision
FROM platform.contract_retirement_provision_v1(
    $1::pg_catalog.text, $2::pg_catalog.text, $3::pg_catalog.text,
    $4::pg_catalog.int8
) AS result(revision);
