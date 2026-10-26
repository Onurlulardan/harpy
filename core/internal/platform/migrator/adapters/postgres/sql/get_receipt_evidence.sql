SELECT environment_id::text,
       release_id::text,
       database_origin,
       database_system_identifier,
       database_oid,
       operation_digest,
       canonical_document,
       canonical_sha256,
       signature_key_id,
       issued_at
FROM platform.bootstrap_receipts
WHERE kind = $1
  AND run_id = $2;
