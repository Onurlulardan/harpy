SELECT platform.billing_installation_assign_v3(
    $1::uuid, $2::uuid, $3::uuid, $4::uuid, $5::uuid, $6::text,
    $7::uuid, $8::bigint, $9::text, $10::timestamptz, $11::bigint,
    $12::text, $13::text, $14::text
);
