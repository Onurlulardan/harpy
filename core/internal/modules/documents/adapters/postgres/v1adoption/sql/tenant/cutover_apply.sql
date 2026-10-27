SELECT applied_count, key_state_sha256, state_sha256
FROM documents_v1_adoption_apply_and_record_v1(
    $1::uuid, $2::uuid, $3::text, $4::uuid, $5::smallint,
    $6::text, $7::text, $8::pg_lsn, $9::pg_lsn, $10::pg_lsn,
    $11::pg_lsn, $12::xid8, $13::integer, $14::text, $15::jsonb,
    $16::bytea, $17::text, $18::bigint, $19::text
)
