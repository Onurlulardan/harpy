SELECT operation_id, state, revision, updated_at
FROM platform.tenancy_onboarding_accept_v2(
    $1::uuid, $2::uuid, $3::text, $4::text, $5::text, $6::text,
    $7::text, $8::text, $9::uuid, $10::name, $11::text, $12::uuid,
    $13::uuid, $14::uuid, $15::uuid, $16::uuid, $17::bigint,
    $18::text, $19::timestamptz, $20::bigint, $21::text, $22::bigint,
    $23::text, $24::uuid, $25::text, $26::uuid, $27::bigint,
    $28::text, $29::text, $30::text, $31::timestamptz
);
