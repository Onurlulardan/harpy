-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_anchor_authority_context_v1(
    p_environment_id uuid,p_tenant_id uuid,p_actor_user_id uuid
)
RETURNS TABLE(
    environment_id uuid,deployment_mode text,owner_binding_id uuid,
    owner_binding_generation bigint,registry_binding_id uuid,head_generation bigint,
    head_state_sha256 bytea,head_ack_sha256 bytea,proposal_signer_generation bigint,
    proposal_signer_kid text,recovery_trust_bundle_sha256 text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NOT platform.iam_platform_current_owner_v1(p_tenant_id,p_actor_user_id) THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='current owner is required';
    END IF;
    RETURN QUERY
    SELECT policy.environment_id,deployment.deployment_mode,policy.owner_binding_id,
        policy.revision,head.registry_binding_id,head.generation,head.state_sha256,
        head.ack_sha256,binding.signer_generation,binding.signer_kid,
        binding.recovery_trust_bundle_sha256
      FROM platform.tenancy_tenant_bindings AS tenant
      JOIN platform.tenancy_deployment_policy AS deployment
        ON deployment.environment_id=tenant.environment_id
      JOIN platform.iam_platform_owner_policies AS policy
        ON policy.environment_id=deployment.environment_id
      JOIN platform.iam_protected_authority_heads AS head
        ON head.environment_id=policy.environment_id
      JOIN platform.platformauthority_bindings AS binding
        ON binding.binding_id=head.registry_binding_id
       AND binding.environment_id=policy.environment_id
       AND binding.deployment_mode=deployment.deployment_mode
       AND binding.signer_status='active'
     WHERE tenant.tenant_id=p_tenant_id AND tenant.environment_id=p_environment_id
       AND tenant.purpose='onpremise_primary' AND deployment.deployment_mode='onpremise';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery anchor authority is unavailable';
    END IF;
END;
$function$;

CREATE FUNCTION platform.iam_recovery_anchor_stage_registration_v1(
    p_operation_id uuid,p_environment_id uuid,p_tenant_id uuid,p_actor_user_id uuid,
    p_idempotency_key text,p_request_sha256 bytea,p_credential_id bytea,
    p_public_key_cose bytea,p_aaguid uuid,p_attestation_format text,
    p_attestation_type text,p_rp_id text,p_allowed_origins jsonb,
    p_allowed_origins_sha256 bytea,p_sign_count bigint,p_occurred_at timestamptz
)
RETURNS TABLE(anchor_id uuid,anchor_set_sha256 bytea,replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_operation platform.iam_recovery_anchor_operations%ROWTYPE;
v_anchor platform.iam_recovery_anchors%ROWTYPE;
v_generation bigint;
v_set bytea;
BEGIN
    IF p_operation_id IS NULL OR p_environment_id IS NULL OR p_tenant_id IS NULL OR
       p_actor_user_id IS NULL OR p_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$' OR
       pg_catalog.octet_length(p_request_sha256)<>32 OR
       pg_catalog.octet_length(p_credential_id) NOT BETWEEN 1 AND 1024 OR
       pg_catalog.octet_length(p_public_key_cose) NOT BETWEEN 1 AND 4096 OR
       p_aaguid IS NULL OR p_aaguid='00000000-0000-0000-0000-000000000000'::uuid OR
       p_attestation_format IN ('','none') OR p_attestation_type IN ('','none') OR
       pg_catalog.octet_length(p_rp_id) NOT BETWEEN 1 AND 253 OR
       pg_catalog.jsonb_typeof(p_allowed_origins)<>'array' OR
       pg_catalog.jsonb_array_length(p_allowed_origins)<1 OR
       pg_catalog.octet_length(p_allowed_origins_sha256)<>32 OR p_sign_count<0 OR
       p_occurred_at IS NULL OR
       NOT platform.iam_platform_current_owner_v1(p_tenant_id,p_actor_user_id) THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='recovery anchor registration is invalid';
    END IF;
    SELECT * INTO v_operation FROM platform.iam_recovery_anchor_operations AS operation
     WHERE operation.operation_id=p_operation_id FOR UPDATE;
    IF NOT FOUND OR v_operation.environment_id<>p_environment_id OR
       v_operation.tenant_id<>p_tenant_id OR v_operation.actor_user_id<>p_actor_user_id OR
       v_operation.kind<>'anchor_register' THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery anchor challenge is unavailable';
    END IF;
    IF v_operation.state<>'options_issued' THEN
        SELECT * INTO v_anchor FROM platform.iam_recovery_anchors AS anchor
         WHERE anchor.anchor_id=v_operation.anchor_id;
        IF v_operation.register_idempotency_key<>p_idempotency_key OR
           v_operation.register_request_sha256<>p_request_sha256 OR
           v_anchor.credential_id<>p_credential_id OR v_anchor.public_key_cose<>p_public_key_cose OR
           v_anchor.aaguid<>p_aaguid OR v_anchor.attestation_format<>p_attestation_format OR
           v_anchor.attestation_type<>p_attestation_type OR v_anchor.rp_id<>p_rp_id OR
           v_anchor.allowed_origins_sha256<>p_allowed_origins_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='23505',MESSAGE='recovery anchor replay conflicts';
        END IF;
        RETURN QUERY SELECT v_anchor.anchor_id,v_operation.anchor_set_sha256,true;
        RETURN;
    END IF;
    IF v_operation.challenge_expires_at<=p_occurred_at THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery anchor challenge expired';
    END IF;
    SELECT COALESCE(MAX(anchor.generation),0)+1 INTO v_generation
      FROM platform.iam_recovery_anchors AS anchor
     WHERE anchor.environment_id=p_environment_id;
    INSERT INTO platform.iam_recovery_anchors(
        anchor_id,environment_id,tenant_id,owner_user_id,generation,credential_id,
        public_key_cose,aaguid,attestation_format,attestation_type,rp_id,allowed_origins,
        allowed_origins_sha256,user_verification,backup_eligible,backup_state,sign_count,
        counter_policy,state,revision,created_at,updated_at
    ) VALUES(v_operation.anchor_id,p_environment_id,p_tenant_id,p_actor_user_id,v_generation,
        p_credential_id,p_public_key_cose,p_aaguid,p_attestation_format,p_attestation_type,
        p_rp_id,p_allowed_origins,p_allowed_origins_sha256,'required',false,false,p_sign_count,
        'adr0062','pending',1,p_occurred_at,p_occurred_at);
    v_set:=platform.iam_recovery_anchor_set_sha256_v1(p_environment_id,NULL);
    UPDATE platform.iam_recovery_anchor_operations
       SET register_idempotency_key=p_idempotency_key,register_request_sha256=p_request_sha256,
           anchor_set_sha256=v_set,state='authority_pending',updated_at=p_occurred_at
     WHERE operation_id=p_operation_id AND state='options_issued';
    IF NOT FOUND THEN RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery anchor changed'; END IF;
    RETURN QUERY SELECT v_operation.anchor_id,v_set,false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_authority_context_v1(uuid,uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_stage_registration_v1(
    uuid,uuid,uuid,uuid,text,bytea,bytea,bytea,uuid,text,text,text,jsonb,bytea,bigint,timestamptz
) FROM PUBLIC;
