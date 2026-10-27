SELECT environment_id::text,tenant_id::text,owner_user_id::text,
       owner_binding_id::text,owner_binding_generation,
       anchor_id::text,anchor_generation,credential_id,public_key_cose,
       aaguid::text,attestation_format,attestation_type,rp_id,
       allowed_origins,allowed_origins_sha256,user_verification,
       backup_eligible,backup_state,sign_count,counter_policy,
       protected_head_generation,protected_state_sha256,protected_ack_sha256
FROM platform.iam_recovery_owner_anchor_evidence_v1($1)
