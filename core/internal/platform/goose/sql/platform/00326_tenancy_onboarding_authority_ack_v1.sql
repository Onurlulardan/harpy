-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_authority_ack_v1(
    in_operation_id uuid,
    in_registry_binding_id uuid,
    in_scope_kind text,
    in_scope_id uuid,
    in_generation bigint,
    in_arg text,
    in_state_sha256 text,
    in_proposal_sha256 text,
    in_ack_sha256 text,
    in_registry_sequence bigint,
    in_verified_at timestamptz,
    in_expected_revision bigint
)
RETURNS TABLE (
    operation_id uuid,
    state text,
    revision bigint,
    updated_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        in_operation_id,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp()
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_authority_ack_v1(uuid, uuid, text, uuid, bigint, text, text, text, text, bigint, timestamptz, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_authority_ack_v1(uuid, uuid, text, uuid, bigint, text, text, text, text, bigint, timestamptz, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_authority_ack_v1(uuid, uuid, text, uuid, bigint, text, text, text, text, bigint, timestamptz, bigint);
