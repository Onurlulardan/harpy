SELECT outcome,replay,id,name,slug,status,lifecycle_revision,settings_revision,
       created_at,updated_at,offboarded_at,purged_at
FROM platform.tenancy_customer_lifecycle_mutate_v1(
    $1::uuid,$2::uuid,$3::uuid,$4::uuid,$5::text,$6::text,$7::bigint,
    $8::text,$9::bytea,$10::timestamptz
)
