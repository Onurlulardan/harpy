SELECT resolved.schema_name, resolved.slug, resolved.status, resolved.not_deleted
FROM platform.tenancy_runtime_resolve_identity_v3($1::uuid) AS resolved;
