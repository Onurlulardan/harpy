SELECT tenant_id,tenant_schema::text,export_id,job_id,attempt_id,
       attempt_generation,lease_owner_id,effect_id,effect_request_sha256,
       unknown_at
FROM platform.platform_async_analytics_dashboard_export_unknown_candidate_v1();
