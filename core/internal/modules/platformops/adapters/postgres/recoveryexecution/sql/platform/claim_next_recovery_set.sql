SELECT attempt_id,lease_token,attempt_generation,recovery_state,
       recovery_set_id,operation_id,environment_id,sequence,
       primary_binding_id,tenant_set_sha256,owner_binding_id,
       owner_binding_generation,database_lineage_id,
       database_system_id,recovery_signing_generation,
       recovery_trust_bundle_sha256,component_inventory,database_now
FROM platform.platformops_claim_next_recovery_set_v3($1,$2)
