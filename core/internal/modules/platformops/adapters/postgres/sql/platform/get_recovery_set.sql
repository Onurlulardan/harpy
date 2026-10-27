SELECT recovery_set_id,operation_id,environment_id,sequence,state,
       manifest_sha256,manifest_bytes,legal_hold,database_system_id,
       recovery_signing_generation,recovery_trust_bundle_sha256,
       protected_at,created_at,updated_at
FROM platform.platformops_get_recovery_set_v1($1)
