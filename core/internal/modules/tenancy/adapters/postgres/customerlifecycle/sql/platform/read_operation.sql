SELECT operation_id,control_tenant_id,target_tenant_id,kind,state,
       result_revision,replay,created_at,updated_at,completed_at,failure_code
FROM platform.tenancy_customer_lifecycle_read_v1($1::uuid,$2::uuid)
