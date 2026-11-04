-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_owner_transfer_compensate_prepare_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_actor_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint,
    in_expected_platform_clock bigint,
    in_purpose_sha256 bytea,
    in_auth_time timestamptz,
    in_acr text,
    in_amr_password boolean,
    in_amr_web_authn boolean,
    in_amrtotp boolean,
    in_original_operation_id uuid,
    in_original_ack_sha256 bytea,
    in_expected_original_generation bigint,
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
AS $fn$
    SELECT
        in_operation_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::bytea,
        ''::bytea,
        in_proposal_sha256,
        ''::text,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_owner_transfer_compensate_prepare_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, bytea, timestamptz, text, boolean, boolean, boolean, uuid, bytea, bigint, bytea, text, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_owner_transfer_compensate_prepare_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, bytea, timestamptz, text, boolean, boolean, boolean, uuid, bytea, bigint, bytea, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_owner_transfer_compensate_prepare_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, bytea, timestamptz, text, boolean, boolean, boolean, uuid, bytea, bigint, bytea, text, bytea, timestamptz);
