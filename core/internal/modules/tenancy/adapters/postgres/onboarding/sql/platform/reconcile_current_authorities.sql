SELECT billing_global_reconciled, billing_global_remaining,
       billing_tenant_reconciled, billing_tenant_remaining,
       workflow_reconciled, workflow_remaining,
       analytics_reconciled, analytics_remaining
FROM platform.tenancy_authority_reconcile_current_v1($1::uuid, $2::text);
