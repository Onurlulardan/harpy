-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_recovery_anchor_authority_context_v1(
    in_environment_id uuid,
    in_tenant_id uuid,
    in_actor_id uuid
)
RETURNS TABLE (
    environment_id uuid,
    deployment_mode text,
    owner_binding_id uuid,
    owner_binding_generation bigint,
    registry_binding_id uuid,
    proposal_signer_generation bigint,
    proposal_signer_kid text,
    recovery_trust_bundle_sha256 text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        in_environment_id,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::text,
        ''::text
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_recovery_anchor_authority_context_v1(uuid, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_recovery_anchor_authority_context_v1(uuid, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_recovery_anchor_authority_context_v1(uuid, uuid, uuid);
