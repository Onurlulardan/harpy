SELECT * FROM audit_entry_append_v1(
    $1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid,
    $6::text, $7::text, $8::text, $9::text, $10::text,
    $11::text, $12::bigint, $13::text, $14::jsonb, $15::bytea,
    $16::timestamptz, $17::timestamptz)
