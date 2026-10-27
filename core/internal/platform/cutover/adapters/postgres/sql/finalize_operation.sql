SELECT terminal_state, terminal_evidence_sha256
FROM cutover_authority.finalize_operation_v1(
    $1::uuid, $2::uuid, $3::uuid, $4::text, $5::bytea,
    $6::text, $7::bytea, $8::text, $9::bytea,
    $10::text, $11::bytea, $12::text, $13::bytea,
    $14::text, $15::text, $16::bytea
)
