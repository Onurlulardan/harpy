SELECT actor_id,token_version,authorization_version,catalog_version_id,
       catalog_sha256,query_document,query_sha256,format,state,attempt_generation
FROM iam_system_record_export_worker_load_v1($1::name,$2,$3,$4,$5)
