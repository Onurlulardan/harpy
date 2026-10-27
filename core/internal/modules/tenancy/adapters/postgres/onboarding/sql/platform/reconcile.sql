SELECT operation_id, state, revision, updated_at
FROM platform.tenancy_onboarding_reconcile_v1(
    $1::uuid, $2::uuid, $3::text, $4::text, $5::bigint
);
