SELECT platform.tenancy_activation_initialize_v5(
    $1::uuid, $2::uuid, $3::uuid, $4::text, $5::text, $6::uuid,
    $7::text, $8::text, $9::text, $10::text
);
