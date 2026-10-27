SELECT operation_id,recovery_set_id,environment_id,target_disposition,state,
       plan_sha256,receipt_sha256,new_database_system_id,
       created_at,updated_at,completed_at
FROM platform.platformops_get_restore_v1($1)
