SELECT binding_id::text,environment_id::text,deployment_mode,registry_kind,
       COALESCE(license_binding_id::text,''),finalization_receipt_sha256,
       recovery_trust_bundle_sha256,signer_generation,signer_kid,
       signer_public_key_sha256,binding_proof_sha256,binding_proof_bytes,
       operation_id::text,purpose,scope_kind,scope_id::text,generation,
       COALESCE(previous_head_sha256,''),state_sha256,state_bytes,
       proposal_sha256,proposal_bytes,ack_sha256,ack_bytes,
       registry_sequence,completed_at
FROM platform.platformauthority_recovery_head_evidence_v1($1)
