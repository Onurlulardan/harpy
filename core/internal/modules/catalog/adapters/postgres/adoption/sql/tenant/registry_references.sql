SELECT 'module', module.id::text, module.id::text, module.slug, module.version, ''
FROM modules AS module
WHERE module.tenant_id = $1::uuid AND module.deleted_at IS NULL
UNION ALL
SELECT 'manifest', schema_version.id::text, '', schema_version.module_slug,
       schema_version.module_version, schema_version.manifest_hash
FROM module_schema_versions AS schema_version
WHERE schema_version.tenant_id = $1::uuid
ORDER BY 1, 2
