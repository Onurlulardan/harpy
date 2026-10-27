SELECT kind,tenant_id,tenant_schema,export_id,authority_group_id,
       authority_group_generation,cutover_generation,job_id,attempt_id,
       attempt_number,lease_generation,original_lease_owner_id,effect_id,
       effect_request_sha256,unknown_at
FROM platform.platform_async_analytics_dashboard_export_maintenance_unknown_candidate_v1();
