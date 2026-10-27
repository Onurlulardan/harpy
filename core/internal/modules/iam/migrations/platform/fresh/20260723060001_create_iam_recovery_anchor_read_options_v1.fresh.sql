-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_anchor_set_sha256_v1(
    p_environment_id uuid,p_excluded_anchor_id uuid DEFAULT NULL
)
RETURNS bytea LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
RETURN (
SELECT pg_catalog.sha256(pg_catalog.convert_to(
    'forgecrud/recovery-anchor-set/v1' || COALESCE(pg_catalog.jsonb_agg(
        pg_catalog.jsonb_build_object(
            'anchor_id',anchor.anchor_id,'generation',anchor.generation,
            'owner_user_id',anchor.owner_user_id,
            'credential_id',pg_catalog.encode(anchor.credential_id,'base64'),
            'public_key_cose',pg_catalog.encode(anchor.public_key_cose,'base64'),
            'aaguid',anchor.aaguid,'rp_id',anchor.rp_id,
            'allowed_origins_sha256',pg_catalog.encode(anchor.allowed_origins_sha256,'hex'),
            'user_verification',anchor.user_verification,
            'backup_eligible',anchor.backup_eligible,'backup_state',anchor.backup_state,
            'sign_count',anchor.sign_count,'counter_policy',anchor.counter_policy
        ) ORDER BY anchor.generation
    ),'[]'::jsonb)::text,'UTF8'))
FROM platform.iam_recovery_anchors AS anchor
WHERE anchor.environment_id=p_environment_id
  AND anchor.state IN ('pending','active')
  AND anchor.anchor_id IS DISTINCT FROM p_excluded_anchor_id
);
END;
$function$;

CREATE FUNCTION platform.iam_recovery_anchor_list_v1(
    p_environment_id uuid,p_tenant_id uuid,p_actor_user_id uuid
)
RETURNS TABLE(anchor_id uuid,generation bigint,credential_id bytea,rp_id text,
              user_verification text,backup_eligible boolean,backup_state boolean,state text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NOT platform.iam_platform_current_owner_v1(p_tenant_id,p_actor_user_id) THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='current owner is required';
    END IF;
    RETURN QUERY SELECT anchor.anchor_id,anchor.generation,anchor.credential_id,
        anchor.rp_id,anchor.user_verification,anchor.backup_eligible,
        anchor.backup_state,anchor.state
      FROM platform.iam_recovery_anchors AS anchor
     WHERE anchor.environment_id=p_environment_id AND anchor.tenant_id=p_tenant_id
     ORDER BY anchor.generation,anchor.anchor_id;
END;
$function$;

CREATE FUNCTION platform.iam_recovery_anchor_options_v1(
    p_operation_id uuid,p_anchor_id uuid,p_environment_id uuid,p_tenant_id uuid,
    p_actor_user_id uuid,p_idempotency_key text,p_request_sha256 bytea,
    p_challenge text,p_rp_id text,p_allowed_origins jsonb,
    p_allowed_origins_sha256 bytea,p_expires_at timestamptz,p_occurred_at timestamptz
)
RETURNS TABLE(operation_id uuid,anchor_id uuid,challenge text,rp_id text,
              allowed_origins jsonb,expires_at timestamptz,replayed boolean)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_existing platform.iam_recovery_anchor_operations%ROWTYPE;
BEGIN
    IF p_operation_id IS NULL OR p_anchor_id IS NULL OR p_environment_id IS NULL OR
       p_tenant_id IS NULL OR p_actor_user_id IS NULL OR
       p_idempotency_key !~ '^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$' OR
       pg_catalog.octet_length(p_request_sha256)<>32 OR
       p_challenge !~ '^[A-Za-z0-9_-]{43}$' OR
       pg_catalog.octet_length(p_rp_id) NOT BETWEEN 1 AND 253 OR
       pg_catalog.jsonb_typeof(p_allowed_origins)<>'array' OR
       pg_catalog.jsonb_array_length(p_allowed_origins)<1 OR
       pg_catalog.octet_length(p_allowed_origins_sha256)<>32 OR
       p_expires_at<=p_occurred_at OR
       NOT platform.iam_platform_current_owner_v1(p_tenant_id,p_actor_user_id) THEN
        RAISE EXCEPTION USING ERRCODE='22023',MESSAGE='recovery anchor options are invalid';
    END IF;
    SELECT * INTO v_existing FROM platform.iam_recovery_anchor_operations
     WHERE operation_id=p_operation_id FOR SHARE;
    IF FOUND THEN
        IF v_existing.environment_id<>p_environment_id OR v_existing.tenant_id<>p_tenant_id OR
           v_existing.actor_user_id<>p_actor_user_id OR v_existing.kind<>'anchor_register' OR
           v_existing.anchor_id<>p_anchor_id OR
           v_existing.options_idempotency_key<>p_idempotency_key OR
           v_existing.options_request_sha256<>p_request_sha256 OR
           v_existing.rp_id<>p_rp_id OR
           v_existing.allowed_origins_sha256<>p_allowed_origins_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='23505',MESSAGE='recovery anchor options conflict';
        END IF;
        RETURN QUERY SELECT v_existing.operation_id,v_existing.anchor_id,
            v_existing.challenge,v_existing.rp_id,v_existing.allowed_origins,
            v_existing.challenge_expires_at,true;
        RETURN;
    END IF;
    INSERT INTO platform.iam_recovery_anchor_operations(
        operation_id,environment_id,tenant_id,actor_user_id,kind,anchor_id,
        options_idempotency_key,options_request_sha256,challenge,challenge_expires_at,
        rp_id,allowed_origins,allowed_origins_sha256,state,created_at,updated_at
    ) VALUES (p_operation_id,p_environment_id,p_tenant_id,p_actor_user_id,
        'anchor_register',p_anchor_id,p_idempotency_key,p_request_sha256,p_challenge,
        p_expires_at,p_rp_id,p_allowed_origins,p_allowed_origins_sha256,
        'options_issued',p_occurred_at,p_occurred_at);
    IF NOT platform.platformops_record_anchor_operation_v1(
        p_operation_id,p_environment_id,p_actor_user_id,p_idempotency_key,
        p_request_sha256,'anchor_register',p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE='40001',MESSAGE='recovery operation projection conflicts';
    END IF;
    RETURN QUERY SELECT p_operation_id,p_anchor_id,p_challenge,p_rp_id,
        p_allowed_origins,p_expires_at,false;
END;
$function$;

CREATE FUNCTION platform.iam_recovery_anchor_registration_v1(
    p_operation_id uuid,p_tenant_id uuid,p_actor_user_id uuid
)
RETURNS TABLE(operation_id uuid,anchor_id uuid,challenge text,rp_id text,
              allowed_origins jsonb,allowed_origins_sha256 bytea,expires_at timestamptz,state text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NOT platform.iam_platform_current_owner_v1(p_tenant_id,p_actor_user_id) THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='current owner is required';
    END IF;
    RETURN QUERY SELECT operation.operation_id,operation.anchor_id,operation.challenge,
        operation.rp_id,operation.allowed_origins,operation.allowed_origins_sha256,
        operation.challenge_expires_at,operation.state
      FROM platform.iam_recovery_anchor_operations AS operation
     WHERE operation.operation_id=p_operation_id AND operation.tenant_id=p_tenant_id
       AND operation.actor_user_id=p_actor_user_id AND operation.kind='anchor_register';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P0002',MESSAGE='recovery anchor operation is missing';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_set_sha256_v1(uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_list_v1(uuid,uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_options_v1(
    uuid,uuid,uuid,uuid,uuid,text,bytea,text,text,jsonb,bytea,timestamptz,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_registration_v1(uuid,uuid,uuid)
FROM PUBLIC;
