SELECT environment_id,generation,bundle_sha256,bundle_bytes,created_at
FROM platform.platformops_get_signing_bundle_v1($1,$2)
