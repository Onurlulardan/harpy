SELECT target_found, ready, operation_id, license_id,
       licensee_organization_id, environment_id, delivery_kind,
       target_revision, retry_at, database_now
FROM platform.billing_license_refresh_target_acquire_v1();
