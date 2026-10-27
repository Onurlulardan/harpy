WITH staged AS MATERIALIZED (
    SELECT * FROM platform.iam_recovery_anchor_stage_registration_v1(
        $1::uuid,$2::uuid,$3::uuid,$4::uuid,$5::text,$6::bytea,$7::bytea,$8::bytea,
        $9::uuid,$10::text,$11::text,$12::text,$13::jsonb,$14::bytea,$15::bigint,$16::timestamptz
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
