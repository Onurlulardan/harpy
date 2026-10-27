SELECT source_identifier_id, lookup_key_id, identifier_hmac,
       user_id, source_revision, identifier_kind, state, source_sha256
FROM platform.iam_maintenance_directory_snapshot_v1($1::uuid)
ORDER BY source_identifier_id, lookup_key_id
