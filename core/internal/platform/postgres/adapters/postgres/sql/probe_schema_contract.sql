SELECT
    contract_row.database_generation,
    contract_row.minimum_runtime_generation,
    contract_row.migration_manifest_sha256,
    contract_row.extension_manifest_sha256,
    contract_row.status,
    contract_row.verified_run_id::text,
    contract_row.revision,
    EXISTS (
        SELECT 1
        FROM platform.release_generations AS compatible_release
        WHERE compatible_release.environment_id = active_release.environment_id
          AND compatible_release.database_origin = active_release.database_origin
          AND compatible_release.database_generation <= contract_row.database_generation
          AND compatible_release.runtime_generation = $1
          AND compatible_release.minimum_database_generation = $2
          AND compatible_release.migration_manifest_sha256 = $3
          AND compatible_release.extension_manifest_sha256 = $4
          AND compatible_release.postgres_image_digest = $5
    )
FROM platform.schema_contracts AS contract_row
JOIN platform.release_generations AS active_release
  ON active_release.migration_run_id = contract_row.verified_run_id
 AND active_release.database_generation = contract_row.database_generation
 AND active_release.minimum_runtime_generation = contract_row.minimum_runtime_generation
 AND active_release.migration_manifest_sha256 = contract_row.migration_manifest_sha256
 AND active_release.extension_manifest_sha256 = contract_row.extension_manifest_sha256
 AND active_release.postgres_image_digest = $5
JOIN platform.environment_identity AS identity_row
  ON identity_row.singleton = TRUE
 AND identity_row.environment_id = active_release.environment_id
 AND identity_row.database_origin = active_release.database_origin
JOIN platform.migration_runs AS run_row
  ON run_row.run_id = active_release.migration_run_id
 AND run_row.release_id = active_release.release_id
 AND run_row.environment_id = active_release.environment_id
 AND run_row.database_origin = active_release.database_origin
 AND run_row.command = 'up'
 AND run_row.state = 'succeeded'
 AND run_row.migration_manifest_sha256 = active_release.migration_manifest_sha256
 AND run_row.extension_manifest_sha256 = active_release.extension_manifest_sha256
JOIN platform.bootstrap_receipts AS bootstrap_row
  ON bootstrap_row.kind = active_release.bootstrap_receipt_kind
 AND bootstrap_row.run_id = active_release.bootstrap_receipt_run_id
 AND bootstrap_row.release_id = active_release.release_id
 AND bootstrap_row.environment_id = active_release.environment_id
 AND bootstrap_row.database_origin = active_release.database_origin
 AND bootstrap_row.database_system_identifier = identity_row.database_system_identifier
 AND bootstrap_row.database_oid = identity_row.database_oid
JOIN platform.bootstrap_receipts AS extension_row
  ON extension_row.kind = active_release.extension_receipt_kind
 AND extension_row.run_id = active_release.extension_receipt_run_id
 AND extension_row.release_id = active_release.release_id
 AND extension_row.environment_id = active_release.environment_id
 AND extension_row.database_origin = active_release.database_origin
 AND extension_row.database_system_identifier = identity_row.database_system_identifier
 AND extension_row.database_oid = identity_row.database_oid
JOIN platform.bootstrap_receipts AS cleanup_row
  ON cleanup_row.kind = active_release.cleanup_receipt_kind
 AND cleanup_row.run_id = active_release.cleanup_receipt_run_id
 AND cleanup_row.release_id = active_release.release_id
 AND cleanup_row.environment_id = active_release.environment_id
 AND cleanup_row.database_origin = active_release.database_origin
 AND cleanup_row.database_system_identifier = identity_row.database_system_identifier
 AND cleanup_row.database_oid = identity_row.database_oid
WHERE contract_row.component = 'core';
