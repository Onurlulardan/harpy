SELECT operation_id, state, revision, updated_at
FROM platform.tenancy_onboarding_record_v1(
    $1::uuid, $2::uuid, $3::uuid, $4::text, $5::text, $6::text,
    $7::timestamptz, $8::bigint
);
