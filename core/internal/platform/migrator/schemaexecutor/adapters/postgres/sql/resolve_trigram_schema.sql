SELECT registry.schema_name
FROM platform.schema_contracts AS contract
JOIN platform.extension_registry AS registry
  ON registry.name = 'pg_trgm'
 AND registry.capabilities = ARRAY['trigram_search']::text[]
JOIN pg_catalog.pg_extension AS extension
  ON extension.extname = registry.name
 AND extension.extversion = registry.version
JOIN pg_catalog.pg_namespace AS namespace
  ON namespace.oid = extension.extnamespace
 AND namespace.nspname = registry.schema_name
JOIN pg_catalog.pg_roles AS owner_role
  ON owner_role.oid = extension.extowner
 AND owner_role.rolname = registry.owner_role
WHERE contract.component = 'core'
  AND contract.status = 'ready'
ORDER BY (registry.verified_run_id = contract.verified_run_id) DESC,
         registry.registered_at DESC
LIMIT 1;
