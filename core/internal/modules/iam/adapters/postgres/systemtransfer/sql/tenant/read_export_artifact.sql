SELECT export_id,job_id,resource_kind,format,state,processed_rows,
       plaintext_bytes,created_at,ready_at,expires_at,object_key,object_version,
       header_sha256,ciphertext_sha256,ciphertext_bytes,key_id,wrapped_dek,
       wrap_nonce,attempt_generation
FROM iam_system_record_export_content_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,$6
)
