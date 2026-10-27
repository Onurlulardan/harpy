SELECT environment_id::text,
       database_generation,
       migration_manifest_sha256,
       extension_manifest_sha256
FROM platform.database_generation_manifests
WHERE environment_id = $1
  AND database_generation = $2;
