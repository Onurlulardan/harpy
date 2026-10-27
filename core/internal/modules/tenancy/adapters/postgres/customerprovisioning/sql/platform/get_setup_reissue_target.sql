SELECT operation_id,tenant_id,requested_by_subject_id,state,
       setup_delivery_state,failure_code,failure_retryable,correlation_id,
       created_at,updated_at,setup_reissue_operation_id,
       schema_name,first_user_id,first_admin_email
FROM platform.tenancy_customer_setup_reissue_target_v1($1::uuid)
