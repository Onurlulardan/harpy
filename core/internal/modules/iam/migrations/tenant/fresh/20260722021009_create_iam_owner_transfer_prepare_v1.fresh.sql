-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_owner_transfer_prepare_v1(
    p_scope_schema name, p_operation_id uuid, p_request_sha256 bytea,
    p_actor_id uuid, p_actor_revision bigint, p_actor_token_version bigint,
    p_actor_authorization_version bigint, p_expected_platform_clock bigint,
    p_purpose_sha256 bytea, p_auth_time timestamptz, p_acr text,
    p_amr_password boolean, p_amr_webauthn boolean, p_amr_totp boolean,
    p_successor_id uuid, p_successor_revision bigint, p_successor_token_version bigint,
    p_successor_authorization_version bigint, p_successor_subject_id uuid,
    p_successor_binding_id uuid, p_registry_binding_id uuid,
    p_expected_policy_revision bigint, p_expected_head_generation bigint,
    p_expected_head_sha256 bytea, p_state_sha256 bytea, p_proposal_sha256 bytea,
    p_proposal_signer_kid text, p_proposal_signature_sha256 bytea,
    p_proposal_expires_at timestamptz
)
RETURNS TABLE (
    operation_id uuid, registry_binding_id uuid, scope_id uuid, generation bigint,
    previous_head_sha256 bytea, state_sha256 bytea, proposal_sha256 bytea,
    status text, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE v_tenant_id uuid; v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_operation_id IS NULL OR pg_catalog.octet_length(p_request_sha256) <> 32 OR
       p_actor_id IS NULL OR p_successor_id IS NULL OR p_actor_id = p_successor_id OR
       p_actor_revision < 1 OR p_actor_token_version < 1 OR p_actor_authorization_version < 1 OR
       p_successor_revision < 1 OR p_successor_token_version < 1 OR
       p_successor_authorization_version < 1 OR pg_catalog.octet_length(p_purpose_sha256) <> 32 OR
       p_acr <> 'urn:forgecrud:acr:step-up' OR
       NOT (p_amr_password OR p_amr_webauthn OR p_amr_totp) OR
       p_auth_time > v_now OR p_auth_time < v_now - INTERVAL '10 minutes' OR
       p_purpose_sha256 <> pg_catalog.sha256(pg_catalog.convert_to(
           'platform.owner.transfer' || E'\n' || p_operation_id::text || E'\n' ||
           p_successor_id::text || E'\n' || (p_expected_head_generation + 1)::text || E'\n' ||
           pg_catalog.encode(p_state_sha256, 'hex'), 'UTF8')) THEN
        RAISE EXCEPTION USING ERRCODE = '22023', MESSAGE = 'Owner transfer input is invalid';
    END IF;
    SELECT identity.tenant_id INTO v_tenant_id
    FROM tenancy_schema_identity AS identity WHERE identity.singleton;
    PERFORM 1 FROM iam_users AS principal
    WHERE principal.id = p_actor_id AND principal.status = 'active'
      AND principal.identity_kind = 'human' AND principal.revision = p_actor_revision
      AND principal.token_version = p_actor_token_version
      AND principal.authorization_version = p_actor_authorization_version FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001', MESSAGE = 'Owner transfer actor is stale';
    END IF;
    PERFORM 1 FROM iam_users AS principal
    WHERE principal.id = p_successor_id AND principal.status = 'active'
      AND principal.identity_kind = 'human' AND principal.revision = p_successor_revision
      AND principal.token_version = p_successor_token_version
      AND principal.authorization_version = p_successor_authorization_version FOR UPDATE;
    IF NOT FOUND OR v_tenant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P1404', MESSAGE = 'Owner successor is not active';
    END IF;
    RETURN QUERY SELECT result.operation_id,result.registry_binding_id,result.scope_id,
        result.generation,result.previous_head_sha256,result.state_sha256,
        result.proposal_sha256,result.status,result.replayed
    FROM platform.iam_owner_transfer_prepare_v1(
        v_tenant_id,p_operation_id,p_request_sha256,p_actor_id,p_actor_revision,
        p_actor_token_version,p_actor_authorization_version,p_successor_id,p_successor_revision,
        p_successor_token_version,p_successor_authorization_version,p_expected_platform_clock,
        p_successor_subject_id,p_successor_binding_id,p_registry_binding_id,
        p_expected_policy_revision,p_expected_head_generation,p_expected_head_sha256,
        p_state_sha256,p_proposal_sha256,p_proposal_signer_kid,
        p_proposal_signature_sha256,p_proposal_expires_at
    ) AS result;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_owner_transfer_prepare_v1(
    name,uuid,bytea,uuid,bigint,bigint,bigint,bigint,bytea,timestamptz,text,
    boolean,boolean,boolean,uuid,bigint,bigint,bigint,uuid,uuid,uuid,bigint,
    bigint,bytea,bytea,bytea,text,bytea,timestamptz
) FROM PUBLIC;
