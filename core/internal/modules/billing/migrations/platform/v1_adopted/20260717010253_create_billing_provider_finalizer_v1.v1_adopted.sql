-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_finalize_provider_command_v1(
    p_tenant_id uuid,
    p_finalizer_job_id uuid,
    p_finalizer_attempt_id uuid,
    p_finalizer_lease_generation bigint,
    p_finalizer_payload_revision integer,
    p_finalizer_payload_reference text,
    p_finalizer_payload_sha256 text,
    p_command_id uuid,
    p_next_finalizer_job_id uuid,
    p_reconciliation_job_id uuid,
    p_observed_at timestamptz
)
RETURNS TABLE (
    decision text,
    command_id uuid,
    actor_id uuid,
    provider_key text,
    active_job_id uuid,
    finalizer_job_id uuid,
    request_sha256 text,
    state text,
    revision bigint,
    occurred_at timestamptz,
    provider_reference text,
    error_code text
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    stored platform.billing_provider_commands%ROWTYPE;
    authority_row platform.billing_tenant_authorities%ROWTYPE;
    provider_value text;
    expected_kind text;
    active_reference text;
    active_effect_key text;
    active_state text;
    active_attempts integer;
    active_effect_started boolean;
    receipt_status text;
    receipt_reference text;
    active_error text;
    next_state text;
    next_decision text;
    next_error text;
BEGIN
    SELECT command_row.* INTO stored
    FROM platform.billing_provider_commands AS command_row
    JOIN platform.billing_tenant_authorities AS authority
      ON authority.authority_id = command_row.authority_id
     AND authority.tenant_id = command_row.tenant_id
    JOIN platform.billing_provider_accounts AS account
      ON account.id = command_row.provider_account_id
    WHERE command_row.tenant_id = p_tenant_id
      AND command_row.id = p_command_id
    FOR UPDATE OF command_row;
    IF NOT FOUND OR stored.actor_id IS NULL OR stored.active_job_id IS NULL
       OR stored.finalizer_job_id IS DISTINCT FROM p_finalizer_job_id
       OR pg_catalog.uuid_extract_version(p_next_finalizer_job_id) IS DISTINCT FROM 7
       OR pg_catalog.uuid_extract_version(p_reconciliation_job_id)
          IS DISTINCT FROM 7
       OR p_next_finalizer_job_id IN (p_finalizer_job_id, p_reconciliation_job_id)
       OR p_next_finalizer_job_id = stored.active_job_id
       OR p_reconciliation_job_id IN (p_finalizer_job_id, stored.active_job_id)
       OR NOT pg_catalog.isfinite(p_observed_at)
       OR p_observed_at < stored.requested_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P5731',
            MESSAGE = 'billing provider finalizer target is invalid';
    END IF;
    SELECT authority.* INTO authority_row FROM platform.billing_tenant_authorities AS authority
    WHERE authority.authority_id = stored.authority_id;
    SELECT account.provider_key INTO provider_value FROM platform.billing_provider_accounts AS account
    WHERE account.id = stored.provider_account_id;
    IF stored.state IN ('succeeded','failed') THEN
        RETURN QUERY SELECT 'replay'::text, stored.id, stored.actor_id,
            provider_value, stored.active_job_id, stored.finalizer_job_id,
            pg_catalog.encode(stored.request_sha256, 'hex'), stored.state,
            stored.revision, stored.completed_at,
            stored.provider_receipt_reference, stored.last_error_code;
        RETURN;
    END IF;

    IF stored.state = 'reconcile_required' THEN
        expected_kind := 'billing.provider.reconcile';
        active_reference := pg_catalog.concat(
            'billing-provider-reconciliation/', provider_value, '/', stored.id::text
        );
        active_effect_key := pg_catalog.concat(
            'billing:provider:reconcile:',
            pg_catalog.encode(pg_catalog.sha256(
                pg_catalog.convert_to(provider_value, 'UTF8') || '\x00'::bytea ||
                pg_catalog.convert_to(stored.id::text, 'UTF8')
            ), 'hex')
        );
    ELSE
        expected_kind := 'billing.provider.command';
        active_reference := pg_catalog.concat(
            'billing-provider-command/', provider_value, '/', stored.id::text
        );
        active_effect_key := pg_catalog.concat(
            'billing:provider:command:', provider_value, ':', stored.id::text
        );
    END IF;
    SELECT evidence.job_state, evidence.job_attempts,
           evidence.effect_started, evidence.receipt_status,
           evidence.provider_reference, evidence.error_code
      INTO active_state, active_attempts, active_effect_started,
           receipt_status, receipt_reference, active_error
    FROM platform.platform_async_external_effect_evidence_v1(
        p_tenant_id, 'tenant', p_tenant_id,
        authority_row.authority_group_id,
        authority_row.authority_group_generation,
        authority_row.billing_domain, authority_row.billing_generation,
        p_finalizer_job_id, p_finalizer_attempt_id,
        p_finalizer_lease_generation, 'billing.provider.finalize',
        p_finalizer_payload_revision, p_finalizer_payload_reference,
        p_finalizer_payload_sha256,
        pg_catalog.concat('billing:provider:finalize:', p_command_id::text),
        stored.active_job_id, expected_kind, 1, active_reference,
        pg_catalog.encode(stored.request_sha256, 'hex'),
        active_effect_key, active_effect_key, provider_value,
        CASE expected_kind WHEN 'billing.provider.command'
            THEN 'command' ELSE 'reconcile' END,
        stored.id
    ) AS evidence;
    IF active_state IN ('queued','running') THEN
        UPDATE platform.billing_provider_commands AS command_row
           SET finalizer_job_id = p_next_finalizer_job_id,
               revision = command_row.revision + 1
        WHERE command_row.id = stored.id
        RETURNING command_row.* INTO stored;
        RETURN QUERY SELECT 'pending'::text, stored.id, stored.actor_id,
            provider_value, stored.active_job_id, stored.finalizer_job_id,
            pg_catalog.encode(stored.request_sha256, 'hex'), stored.state,
            stored.revision, p_observed_at, stored.provider_receipt_reference,
            stored.last_error_code;
        RETURN;
    END IF;

    IF receipt_status = 'succeeded' THEN
        next_state := 'succeeded';
        next_decision := 'succeeded';
        next_error := NULL;
    ELSIF receipt_status = 'failed' OR NOT active_effect_started THEN
        next_state := 'failed';
        next_decision := 'failed';
        receipt_reference := NULL;
        next_error := COALESCE(
            active_error, 'PROVIDER_COMMAND_FAILED'
        );
    ELSIF stored.state = 'reconcile_required' THEN
        next_state := 'reconcile_required';
        next_decision := 'reconcile_unknown';
        receipt_reference := NULL;
        next_error := 'JOB_OUTCOME_UNKNOWN';
    ELSE
        next_state := 'reconcile_required';
        next_decision := 'reconcile';
        receipt_reference := NULL;
        next_error := 'JOB_OUTCOME_UNKNOWN';
    END IF;

    UPDATE platform.billing_provider_commands AS command_row
       SET state = next_state,
           active_job_id = CASE WHEN next_decision = 'reconcile'
               THEN p_reconciliation_job_id ELSE command_row.active_job_id END,
           finalizer_job_id = CASE WHEN next_decision = 'reconcile'
               THEN p_next_finalizer_job_id ELSE command_row.finalizer_job_id END,
           provider_receipt_reference = receipt_reference,
           last_error_code = next_error,
           attempts = LEAST(
               25, CASE WHEN command_row.state = 'reconcile_required'
                   THEN command_row.attempts + active_attempts
                   ELSE active_attempts END
           ),
           completed_at = p_observed_at,
           revision = command_row.revision + 1
    WHERE command_row.id = stored.id
    RETURNING command_row.* INTO stored;

    RETURN QUERY SELECT next_decision, stored.id, stored.actor_id,
        provider_value, stored.active_job_id, stored.finalizer_job_id,
        pg_catalog.encode(stored.request_sha256, 'hex'), stored.state,
        stored.revision, stored.completed_at,
        stored.provider_receipt_reference, stored.last_error_code;
END;
$function$;

REVOKE ALL ON FUNCTION platform.billing_finalize_provider_command_v1(
    uuid,uuid,uuid,bigint,integer,text,text,uuid,uuid,uuid,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_finalize_provider_command_v1(
    uuid,uuid,uuid,bigint,integer,text,text,uuid,uuid,uuid,timestamptz
) TO harpy_runtime;
-- +goose StatementEnd
