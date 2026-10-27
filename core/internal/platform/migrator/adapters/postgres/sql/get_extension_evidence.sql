SELECT schema_name,
       owner_role,
       control_file_sha256,
       pg_catalog.array_to_json(capabilities)::text,
       database_origin
FROM platform.extension_registry
WHERE name = $1
  AND version = $2
  AND verified_run_id = $3;
