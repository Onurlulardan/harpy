SELECT migration_run_id::text,
       release_id::text,
       environment_id::text,
       database_origin,
       database_generation,
       minimum_runtime_generation,
       runtime_generation,
       minimum_database_generation,
       migration_manifest_sha256,
       extension_manifest_sha256,
       postgres_image_digest,
       bootstrap_receipt_run_id::text,
       extension_receipt_run_id::text,
       cleanup_receipt_run_id::text
FROM platform.release_generations
WHERE environment_id = $1
  AND release_id = $2;
