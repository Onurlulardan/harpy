SELECT schema_name, slug, status, not_deleted
FROM platform.tenancy_runtime_resolve_command_v1($1);
