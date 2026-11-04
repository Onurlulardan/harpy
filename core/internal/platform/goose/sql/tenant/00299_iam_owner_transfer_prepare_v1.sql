-- +goose Up
CREATE FUNCTION iam_owner_transfer_prepare_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_actor_id uuid,
    in_actor_expected_revision bigint,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint,
    in_expected_platform_clock bigint,
    in_purpose_sha256 bytea,
    in_auth_time timestamptz,
    in_acr text,
    in_amr_password boolean,
    in_amr_web_authn boolean,
    in_amrtotp boolean,
    in_target_user_id uuid,
    in_target_expected_revision bigint,
    in_target_expected_token_version bigint,
    in_target_expected_authorization_version bigint,
    in_successor_subject_id uuid,
    in_successor_binding_id uuid,
    in_registry_binding_id uuid,
    in_expected_owner_policy_revision bigint,
    in_expected_head_generation bigint,
    in_expected_head_sha256 bytea,
    in_state_sha256 bytea,
    in_proposal_sha256 bytea,
    in_proposal_signer_kid text,
    in_proposal_signature_sha256 bytea,
    in_proposal_expires_at timestamptz
)
RETURNS TABLE (
    operation_id uuid,
    registry_binding_id uuid,
    scope_id uuid,
    generation bigint,
    previous_head_sha256 bytea,
    state_sha256 bytea,
    proposal_sha256 bytea,
    status text,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_operation_id,
        in_registry_binding_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::bytea,
        in_state_sha256,
        in_proposal_sha256,
        ''::text,
        false
    ;
$$;

REVOKE ALL ON FUNCTION iam_owner_transfer_prepare_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, bigint, bytea, timestamptz, text, boolean, boolean, boolean, uuid, bigint, bigint, bigint, uuid, uuid, uuid, bigint, bigint, bytea, bytea, bytea, text, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_owner_transfer_prepare_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, bigint, bytea, timestamptz, text, boolean, boolean, boolean, uuid, bigint, bigint, bigint, uuid, uuid, uuid, bigint, bigint, bytea, bytea, bytea, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_owner_transfer_prepare_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, bigint, bytea, timestamptz, text, boolean, boolean, boolean, uuid, bigint, bigint, bigint, uuid, uuid, uuid, bigint, bigint, bytea, bytea, bytea, text, bytea, timestamptz);
