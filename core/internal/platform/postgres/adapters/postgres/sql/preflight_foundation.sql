SELECT
    count(*) OVER ()::bigint,
    identity.database_origin,
    identity.database_system_identifier,
    identity.database_oid,
    identity.environment_id::text,
    identity.release_id::text,
    active_release.release_id::text,
    run.run_id::text,
    run.state,
	run.command,
    run.migration_manifest_sha256,
    run.extension_manifest_sha256,
    contract.database_generation,
    contract.minimum_runtime_generation,
    contract.migration_manifest_sha256,
    contract.extension_manifest_sha256,
    contract.status,
    contract.verified_run_id::text,
    contract.revision,
    extension.version,
    extension.schema_name,
    extension.owner_role,
    extension.database_origin,
    extension.verified_run_id::text,
    EXISTS (
        SELECT 1
        FROM platform.release_generations AS compatible_release
        WHERE compatible_release.environment_id = identity.environment_id
          AND compatible_release.database_origin = identity.database_origin
          AND compatible_release.database_generation <= contract.database_generation
          AND compatible_release.runtime_generation = $2
          AND compatible_release.minimum_database_generation = $3
          AND compatible_release.migration_manifest_sha256 = $4
          AND compatible_release.extension_manifest_sha256 = $5
          AND compatible_release.postgres_image_digest = $6
    )
FROM platform.environment_identity AS identity
JOIN platform.schema_contracts AS contract
  ON contract.component = 'core'
JOIN platform.migration_runs AS run
  ON run.run_id = contract.verified_run_id
 AND run.environment_id = identity.environment_id
 AND run.database_origin = identity.database_origin
JOIN platform.release_generations AS active_release
  ON active_release.migration_run_id = run.run_id
 AND active_release.release_id = run.release_id
 AND active_release.environment_id = run.environment_id
 AND active_release.database_origin = run.database_origin
 AND active_release.database_generation = contract.database_generation
 AND active_release.minimum_runtime_generation = contract.minimum_runtime_generation
 AND active_release.migration_manifest_sha256 = contract.migration_manifest_sha256
 AND active_release.extension_manifest_sha256 = contract.extension_manifest_sha256
 AND active_release.postgres_image_digest = $6
JOIN platform.bootstrap_receipts AS bootstrap_receipt
  ON bootstrap_receipt.kind = active_release.bootstrap_receipt_kind
 AND bootstrap_receipt.run_id = active_release.bootstrap_receipt_run_id
 AND bootstrap_receipt.environment_id = identity.environment_id
 AND bootstrap_receipt.release_id = active_release.release_id
 AND bootstrap_receipt.database_origin = identity.database_origin
 AND bootstrap_receipt.database_system_identifier = identity.database_system_identifier
 AND bootstrap_receipt.database_oid = identity.database_oid
JOIN platform.bootstrap_receipts AS extension_receipt
  ON extension_receipt.kind = active_release.extension_receipt_kind
 AND extension_receipt.run_id = active_release.extension_receipt_run_id
 AND extension_receipt.environment_id = identity.environment_id
 AND extension_receipt.release_id = active_release.release_id
 AND extension_receipt.database_origin = identity.database_origin
 AND extension_receipt.database_system_identifier = identity.database_system_identifier
 AND extension_receipt.database_oid = identity.database_oid
JOIN platform.bootstrap_receipts AS cleanup_receipt
  ON cleanup_receipt.kind = active_release.cleanup_receipt_kind
 AND cleanup_receipt.run_id = active_release.cleanup_receipt_run_id
 AND cleanup_receipt.environment_id = identity.environment_id
 AND cleanup_receipt.release_id = active_release.release_id
 AND cleanup_receipt.database_origin = identity.database_origin
 AND cleanup_receipt.database_system_identifier = identity.database_system_identifier
 AND cleanup_receipt.database_oid = identity.database_oid
JOIN platform.extension_registry AS extension
  ON extension.name = 'vector'
 AND extension.version = $1
 AND extension.verified_run_id = extension_receipt.run_id
 AND extension.database_origin = identity.database_origin
WHERE identity.singleton = TRUE;
