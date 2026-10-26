SELECT retired_contract_key
FROM platform.contract_retirement_hit_v1(
    $1::pg_catalog.jsonb, $2::pg_catalog.text, $3::pg_catalog.text
);
