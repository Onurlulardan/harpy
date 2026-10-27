-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_recovery_head_evidence_v1(
    p_environment_id uuid
)
RETURNS TABLE(
    binding_id uuid,environment_id uuid,deployment_mode text,
    registry_kind text,license_binding_id uuid,
    finalization_receipt_sha256 text,recovery_trust_bundle_sha256 text,
    signer_generation bigint,signer_kid text,signer_public_key_sha256 text,
    binding_proof_sha256 text,binding_proof_bytes bytea,
    operation_id uuid,purpose text,scope_kind text,scope_id uuid,
    generation bigint,previous_head_sha256 text,state_sha256 text,
    state_bytes bytea,proposal_sha256 text,proposal_bytes bytea,
    ack_sha256 text,ack_bytes bytea,registry_sequence bigint,
    completed_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_backup'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'harpy-v2:core-database-backup-executor'
       OR p_environment_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='protected authority recovery evidence denied';
    END IF;
    RETURN QUERY
    SELECT binding.binding_id,binding.environment_id,binding.deployment_mode,
           binding.registry_kind,binding.license_binding_id,
           binding.finalization_receipt_sha256,
           binding.recovery_trust_bundle_sha256,binding.signer_generation,
           binding.signer_kid,binding.signer_public_key_sha256,
           binding.proof_sha256,binding.proof_bytes,operation.operation_id,
           operation.purpose,operation.scope_kind,operation.scope_id,
           operation.generation,operation.previous_head_sha256,
           operation.state_sha256,operation.state_bytes,
           operation.proposal_sha256,operation.proposal_bytes,
           operation.ack_sha256,operation.ack_bytes,
           operation.registry_sequence,operation.completed_at
    FROM platform.platformauthority_bindings AS binding
    JOIN platform.platformauthority_heads AS head
      ON head.binding_id=binding.binding_id
     AND head.scope_kind='environment' AND head.scope_id=binding.environment_id
    JOIN platform.platformauthority_operations AS operation
      ON operation.binding_id=head.binding_id
     AND operation.scope_kind=head.scope_kind
     AND operation.scope_id=head.scope_id
     AND operation.generation=head.generation
     AND operation.state_sha256=head.state_sha256
     AND operation.proposal_sha256=head.proposal_sha256
     AND operation.ack_sha256=head.ack_sha256
     AND operation.registry_sequence=head.registry_sequence
     AND operation.status='completed'
    WHERE binding.environment_id=p_environment_id
      AND binding.deployment_mode='onpremise'
      AND binding.registry_kind = 'development'
      AND binding.signer_status='active';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P0002',
            MESSAGE='current protected authority evidence is unavailable';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformauthority_recovery_head_evidence_v1(uuid)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformauthority_recovery_head_evidence_v1(uuid)
TO harpy_backup_reader;
