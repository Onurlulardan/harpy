SELECT operation_id, attempt_id, attempt_number, revision, lease_until
FROM platform.tenancy_onboarding_claim_v2(
    $1::uuid, $2::uuid, $3::text, $4::text, $5::bigint, $6::integer
);
