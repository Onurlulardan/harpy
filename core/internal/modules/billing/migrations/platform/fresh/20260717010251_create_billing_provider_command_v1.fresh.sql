-- +goose Up
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
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    stored platform.billing_provider_commands%ROWTYPE;
    resolved_provider_id uuid;
    resolved_provider_key text;
BEGIN
    PERFORM platform.billing_assert_tenant_authority_v1(p_authority_id);
    IF NOT EXISTS (
        SELECT 1
        FROM platform.billing_tenant_authorities AS authority_row
        WHERE authority_row.authority_id = p_authority_id
          AND authority_row.tenant_id = p_tenant_id
    ) OR pg_catalog.uuid_extract_version(p_command_id) <> 7
       OR pg_catalog.uuid_extract_version(p_finalizer_job_id) <> 7
       OR p_command_id = p_finalizer_job_id
       OR p_actor_id = '00000000-0000-0000-0000-000000000000'::uuid
       OR p_subscription_id = '00000000-0000-0000-0000-000000000000'::uuid
       OR p_command_kind NOT IN ('checkout','change','cancel','refund','reconcile')
       OR pg_catalog.octet_length(p_effect_key) NOT BETWEEN 1 AND 256
       OR p_effect_key IS DISTINCT FROM pg_catalog.btrim(p_effect_key)
       OR pg_catalog.octet_length(p_request_sha256) <> 32
       OR NOT pg_catalog.isfinite(p_requested_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing provider command input is invalid';
    END IF;

    SELECT subscription.provider_account_id, account.provider_key
      INTO resolved_provider_id, resolved_provider_key
    FROM platform.billing_subscriptions AS subscription
    JOIN platform.billing_provider_accounts AS account
      ON account.id = subscription.provider_account_id
    WHERE subscription.tenant_id = p_tenant_id
      AND subscription.id = p_subscription_id
      AND account.lifecycle = 'active'
    FOR SHARE OF subscription, account;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P5731',
            MESSAGE = 'billing provider command target is invalid';
    END IF;
    IF p_provider_object_mapping_id IS NOT NULL AND NOT EXISTS (
        SELECT 1
        FROM platform.billing_provider_objects AS mapping
        WHERE mapping.tenant_id = p_tenant_id
          AND mapping.id = p_provider_object_mapping_id
          AND mapping.provider_account_id = resolved_provider_id
          AND mapping.retired_at IS NULL
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5731',
            MESSAGE = 'billing provider command mapping is invalid';
    END IF;

    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'billing:provider-command:' || resolved_provider_id::text ||
        ':' || p_effect_key, 0
    ));
    SELECT * INTO stored
    FROM platform.billing_provider_commands AS command_row
    WHERE command_row.provider_account_id = resolved_provider_id
      AND command_row.effect_key = p_effect_key
    FOR UPDATE;
    IF FOUND THEN
        IF stored.tenant_id IS DISTINCT FROM p_tenant_id
           OR stored.actor_id IS DISTINCT FROM p_actor_id
           OR stored.provider_object_mapping_id IS DISTINCT FROM
              p_provider_object_mapping_id
           OR stored.subscription_id IS DISTINCT FROM p_subscription_id
           OR stored.command_kind IS DISTINCT FROM p_command_kind
           OR stored.request_sha256 IS DISTINCT FROM p_request_sha256
           OR stored.finalizer_job_id IS NULL THEN
            RAISE EXCEPTION USING ERRCODE = 'P5730',
                MESSAGE = 'billing provider command key was reused';
        END IF;
        RETURN QUERY SELECT stored.id, stored.finalizer_job_id, stored.tenant_id,
            stored.actor_id, stored.provider_account_id,
            stored.provider_object_mapping_id, stored.subscription_id,
            resolved_provider_key, stored.command_kind, stored.effect_key,
            stored.request_sha256, stored.state, stored.requested_at,
            stored.revision, true;
        RETURN;
    END IF;

    INSERT INTO platform.billing_provider_commands (
        id, authority_id, tenant_id, provider_account_id,
        provider_object_mapping_id, subscription_id, command_kind,
        effect_key, request_sha256, state, provider_receipt_reference,
        last_error_code, attempts, requested_at, completed_at, revision,
        actor_id, active_job_id, finalizer_job_id
    ) VALUES (
        p_command_id, p_authority_id, p_tenant_id, resolved_provider_id,
        p_provider_object_mapping_id, p_subscription_id, p_command_kind,
        p_effect_key, p_request_sha256, 'queued', NULL, NULL, 0,
        p_requested_at, NULL, 1, p_actor_id, p_command_id, p_finalizer_job_id
    );
    RETURN QUERY SELECT p_command_id, p_finalizer_job_id, p_tenant_id,
        p_actor_id, resolved_provider_id, p_provider_object_mapping_id,
        p_subscription_id, resolved_provider_key, p_command_kind, p_effect_key,
        p_request_sha256, 'queued'::text, p_requested_at, 1::bigint, false;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_create_provider_command_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,text,text,bytea,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_create_provider_command_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,text,text,bytea,timestamptz
) TO forgecrud_runtime;
-- +goose StatementEnd
