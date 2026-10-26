SELECT artifact_bytes,artifact_sha256,expires_at
FROM platform.platformops_claim_recovery_kit_v1($1,$2,$3,$4)
