SELECT source_id, actor_id, resource_kind, object_key, object_version,
       cleanup_attempt, lease_generation
FROM iam_system_import_source_cleanup_claim_v1($1::name, $2, $3, $4)
