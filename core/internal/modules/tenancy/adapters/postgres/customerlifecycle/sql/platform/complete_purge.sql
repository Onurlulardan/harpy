SELECT operation_id,control_tenant_id,target_tenant_id,kind,state,
       result_revision,false,created_at,updated_at,completed_at,failure_code
FROM platform.tenancy_customer_lifecycle_complete_purge_v1(
    $1::uuid,$2::uuid,$3::uuid,$4::bytea,$5::timestamptz
)
