-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_owner_transfer_ack_v1(
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
    in_registry_binding_id uuid,
    in_scope_id uuid,
    in_registry_sequence bigint,
    in_generation bigint,
    in_proposal_sha256 bytea,
    in_previous_head_sha256 bytea,
    in_state_sha256 bytea,
    in_ack_sha256 bytea,
    in_issued_at timestamptz,
    in_registry_kid text
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
        in_registry_binding_id,
        in_scope_id,
        in_generation,
        in_previous_head_sha256,
        in_state_sha256,
        in_proposal_sha256,
        ''::text,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_owner_transfer_ack_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, bytea, timestamptz, text, boolean, boolean, boolean, uuid, uuid, bigint, bigint, bytea, bytea, bytea, bytea, timestamptz, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_owner_transfer_ack_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, bytea, timestamptz, text, boolean, boolean, boolean, uuid, uuid, bigint, bigint, bytea, bytea, bytea, bytea, timestamptz, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_owner_transfer_ack_v1(name, uuid, bytea, uuid, bigint, bigint, bigint, bytea, timestamptz, text, boolean, boolean, boolean, uuid, uuid, bigint, bigint, bytea, bytea, bytea, bytea, timestamptz, text);
