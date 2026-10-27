WITH staged AS MATERIALIZED (
    SELECT * FROM platform.iam_recovery_anchor_stage_delete_v1(
        $1::uuid,$2::uuid,$3::uuid,$4::uuid,$5::uuid,$6::text,$7::bytea,$8::timestamptz
    )
), authority AS MATERIALIZED (
    SELECT * FROM platform.iam_recovery_anchor_authority_context_v1($2::uuid,$3::uuid,$4::uuid)
)
SELECT staged.anchor_id,staged.anchor_set_sha256,staged.replayed,
       authority.environment_id,authority.deployment_mode,authority.owner_binding_id,
       authority.owner_binding_generation,authority.registry_binding_id,
       authority.proposal_signer_generation,authority.proposal_signer_kid,
       authority.recovery_trust_bundle_sha256
FROM staged CROSS JOIN authority;
