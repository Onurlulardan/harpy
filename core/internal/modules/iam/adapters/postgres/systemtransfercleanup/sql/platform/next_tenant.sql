SELECT tenant_id, tenant_schema::text
FROM platform.platform_async_iam_import_cleanup_tenant_v1($1)
