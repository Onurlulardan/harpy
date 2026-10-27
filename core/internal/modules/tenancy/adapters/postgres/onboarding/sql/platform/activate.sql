SELECT operation_id, tenant_id, operation_kind, step_plan_sha256,
       migration_receipt_id, migration_receipt_sha256, authority_ack_sha256,
       revision, database_now, lease_until
FROM platform.tenancy_onboarding_activate_v3($1::uuid, $2::text, $3::bigint);
