SELECT actor_id,resource_kind,authorization_version,duplicate_action,
       mapping_document,next_row_index,processed_rows,source_id,source_format,
       object_key,object_version,header_sha256,ciphertext_sha256,
       ciphertext_bytes,plaintext_sha256,plaintext_bytes,key_id,wrapped_dek,
       wrap_nonce,provider_request_id,uploaded_at,finished,catalog_version_id,
       catalog_sha256,identity_manifest_sha256
FROM iam_system_record_import_worker_claim_v2(
    $1::name,$2,$3,$4,$5,$6,$7
)
