-- +goose Up
-- BELIRSIZ: provider account, provider key, state, revision, and replay are not determined by the caller.
-- +goose StatementBegin
CREATE FUNCTION platform.billing_create_provider_command_v1(
    p_authority_id uuid,
    p_tenant_id uuid,
    p_command_id uuid,
    p_finalizer_job_id uuid,
    p_actor_id uuid,
    p_provider_object_mapping_id uuid,
    p_subscription_id uuid,
    p_command_kind text,
    p_effect_key text,
    p_request_sha256 bytea,
    p_requested_at timestamptz
)
RETURNS TABLE (
    command_id uuid,
    finalizer_job_id uuid,
    tenant_id uuid,
    actor_id uuid,
    provider_account_id uuid,
    provider_object_mapping_id uuid,
    subscription_id uuid,
    provider_key text,
    command_kind text,
    effect_key text,
    request_sha256 bytea,
    state text,
    requested_at timestamptz,
    revision bigint,
    replay boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        p_command_id,
        p_finalizer_job_id,
        p_tenant_id,
        p_actor_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        p_provider_object_mapping_id,
        p_subscription_id,
        ''::text,
        p_command_kind,
        p_effect_key,
        p_request_sha256,
        ''::text,
        p_requested_at,
        0::bigint,
        false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_create_provider_command_v1(
    uuid, uuid, uuid, uuid, uuid, uuid,
    uuid, text, text, bytea, timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_create_provider_command_v1(
    uuid, uuid, uuid, uuid, uuid, uuid,
    uuid, text, text, bytea, timestamptz
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.billing_create_provider_command_v1(
    uuid, uuid, uuid, uuid, uuid, uuid,
    uuid, text, text, bytea, timestamptz
);
