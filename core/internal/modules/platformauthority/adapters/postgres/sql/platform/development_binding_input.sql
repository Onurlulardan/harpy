SELECT operation_id, registry_binding_id,
       finalization_receipt_sha256, recovery_trust_bundle_sha256
FROM platform.platformauthority_development_binding_input_v1($1)
