SELECT resolved.id, resolved.status, resolved.not_deleted
FROM platform.tenancy_runtime_resolve_slug_v1($1::text) AS resolved;
