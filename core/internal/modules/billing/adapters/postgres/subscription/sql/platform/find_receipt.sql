SELECT operation.id, operation.subscription_id, operation.source_operation_id,
       operation.source_kind,
       operation.operation_kind, operation.request_sha256,
       operation.result_sha256, operation.result_provider_account_id,
       operation.result_offering_version_id, operation.result_price_version_id,
       operation.result_currency::text, operation.result_state,
       operation.result_period_start, operation.result_period_end,
       operation.result_trial_ends_at, operation.result_grace_ends_at,
       operation.result_cancel_at_period_end, operation.result_cancel_at,
       operation.result_revision, operation.result_updated_at,
       operation.completed_at
FROM platform.billing_subscription_operations AS operation
WHERE operation.authority_id = $1 AND operation.tenant_id = $2
  AND operation.source_operation_id = $3;
