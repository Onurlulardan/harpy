-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_owner_anchor_evidence_v1(
    p_environment_id uuid
)
RETURNS TABLE(
    environment_id uuid,tenant_id uuid,owner_user_id uuid,
    owner_binding_id uuid,owner_binding_generation bigint,
    anchor_id uuid,anchor_generation bigint,credential_id bytea,
    public_key_cose bytea,aaguid uuid,attestation_format text,
    attestation_type text,rp_id text,allowed_origins jsonb,
    allowed_origins_sha256 bytea,user_verification text,
    backup_eligible boolean,backup_state boolean,sign_count bigint,
    counter_policy text,protected_head_generation bigint,
    protected_state_sha256 bytea,protected_ack_sha256 bytea
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_backup'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'forgecrud-v2:core-database-backup-executor'
       OR p_environment_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery owner evidence denied';
    END IF;
    RETURN QUERY
    SELECT anchor.environment_id,anchor.tenant_id,anchor.owner_user_id,
           policy.owner_binding_id,policy.revision,anchor.anchor_id,
           anchor.generation,anchor.credential_id,anchor.public_key_cose,
           anchor.aaguid,anchor.attestation_format,anchor.attestation_type,
           anchor.rp_id,anchor.allowed_origins,anchor.allowed_origins_sha256,
           anchor.user_verification,anchor.backup_eligible,
           anchor.backup_state,anchor.sign_count,anchor.counter_policy,
           head.generation,head.state_sha256,head.ack_sha256
    FROM platform.iam_recovery_anchors AS anchor
    JOIN platform.iam_platform_owner_policies AS policy
      ON policy.environment_id=anchor.environment_id
     AND policy.revision>0
    JOIN platform.iam_protected_authority_heads AS head
      ON head.environment_id=anchor.environment_id
     AND head.generation=anchor.authority_generation
     AND head.state_sha256=anchor.protected_state_sha256
     AND head.ack_sha256=anchor.authority_ack_sha256
    WHERE anchor.environment_id=p_environment_id
      AND anchor.state='active' AND anchor.user_verification='required'
      AND NOT anchor.backup_eligible AND NOT anchor.backup_state
    ORDER BY anchor.generation,anchor.anchor_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P0002',
            MESSAGE='current recovery owner evidence is unavailable';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_recovery_owner_anchor_evidence_v1(uuid)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_recovery_owner_anchor_evidence_v1(uuid)
TO forgecrud_backup_reader;
