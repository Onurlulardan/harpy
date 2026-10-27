SELECT operation_id,tenant_id,requested_by_subject_id,state,
       setup_delivery_state,failure_code,failure_retryable,correlation_id,
       created_at,updated_at
FROM platform.tenancy_customer_provision_read_v1($1::uuid,$2::uuid)
