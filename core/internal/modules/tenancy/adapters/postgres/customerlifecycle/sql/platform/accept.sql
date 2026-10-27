SELECT outcome,operation_id,control_tenant_id,target_tenant_id,kind,state,
       result_revision,replay,created_at,updated_at,completed_at,failure_code
FROM platform.tenancy_customer_lifecycle_accept_v1(
    $1::uuid,$2::uuid,$3::uuid,$4::uuid,$5::text,$6::bigint,$7::text,
    $8::bytea,$9::uuid,$10::uuid,$11::uuid,$12::bytea,$13::timestamptz,
    $14::timestamptz
)
