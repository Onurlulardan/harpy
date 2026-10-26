SELECT database_generation,
       minimum_runtime_generation,
       migration_manifest_sha256,
       extension_manifest_sha256,
       status,
       verified_run_id::text,
       revision
FROM platform.schema_contracts
WHERE component = 'core';
