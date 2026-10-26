SELECT * FROM audit_entries_list_v1(
    $1::text, $2::text, $3::text, $4::text, $5::text,
    $6::uuid, $7::uuid, $8::uuid, $9::timestamptz, $10::timestamptz,
    $11::text, $12::integer, $13::integer)
