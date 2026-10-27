SELECT operation_id, state, revision, updated_at
FROM platform.tenancy_onboarding_authority_ack_v1(
    $1::uuid, $2::uuid, $3::text, $4::uuid, $5::bigint, $6::text,
    $7::text, $8::text, $9::text, $10::bigint, $11::timestamptz, $12::bigint
);
