-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_apply_v1_base_assignment_v1(
    p_operation_id uuid, p_run_id uuid, p_record jsonb
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_id uuid;
    v_sha bytea;
BEGIN
    v_id := (p_record->>'id')::uuid;
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
      AND run.scope_kind = 'tenant' AND run.state = 'planned' AND run.eligible;
    IF (p_record->>'tenant_id')::uuid IS DISTINCT FROM v_run.scope_id THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing V1 base assignment tenant is invalid';
    END IF;
    INSERT INTO platform.billing_tenant_base_offerings (
        id, authority_id, tenant_id, offering_version_id,
        effective_start, effective_end, source_operation_id, created_at
    ) VALUES (
        v_id, v_run.authority_id, v_run.scope_id,
        (p_record->>'offering_version_id')::uuid,
        (p_record->>'effective_start')::timestamptz, NULL,
        (p_record->>'source_operation_id')::uuid, v_run.snapshot_at
    );
    v_sha := platform.billing_v1_adoption_tenant_fingerprint_v1(
        'base_assignment', v_id
    );
    PERFORM platform.billing_record_v1_adoption_mapping_v1(
        p_operation_id, p_run_id, 'base_assignments', v_run.scope_id::text,
        'base_assignment', v_id, v_sha, v_run.snapshot_at
    );
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_apply_v1_base_assignment_v1(
    uuid, uuid, jsonb
) FROM PUBLIC;

CREATE FUNCTION platform.billing_apply_v1_subscription_v1(
    p_operation_id uuid, p_run_id uuid, p_record jsonb
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_id uuid;
    v_created timestamptz;
    v_sha bytea;
BEGIN
    v_id := (p_record->>'id')::uuid;
    v_created := (p_record->>'created_at')::timestamptz;
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
      AND run.scope_kind = 'tenant' AND run.state = 'planned' AND run.eligible;
    IF (p_record->>'tenant_id')::uuid IS DISTINCT FROM v_run.scope_id THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing V1 subscription tenant is invalid';
    END IF;
    INSERT INTO platform.billing_subscriptions (
        id, authority_id, tenant_id, provider_account_id,
        offering_version_id, price_version_id, state,
        period_start, period_end, trial_ends_at, grace_ends_at,
        cancel_at_period_end, cancel_at, created_at, updated_at, revision
    ) VALUES (
        v_id, v_run.authority_id, v_run.scope_id,
        (p_record->>'provider_account_id')::uuid,
        (p_record->>'offering_version_id')::uuid,
        (p_record->>'price_version_id')::uuid, p_record->>'state',
        (p_record->>'period_start')::timestamptz,
        (p_record->>'period_end')::timestamptz,
        NULLIF(p_record->>'trial_ends_at', '')::timestamptz,
        NULLIF(p_record->>'grace_ends_at', '')::timestamptz,
        (p_record->>'cancel_at_period_end')::boolean,
        NULLIF(p_record->>'cancel_at', '')::timestamptz,
        v_created, v_created, 1
    );
    INSERT INTO platform.billing_v1_subscription_wire_archive (
        subscription_id, authority_id, tenant_id, source_subscription_id,
        source_updated_at, trial_started_at, cancelled_at, cancellation_reason,
        adoption_operation_id, adoption_run_id, archived_at
    ) VALUES (
        v_id, v_run.authority_id, v_run.scope_id,
        (p_record->>'source_id')::uuid,
        (p_record->>'updated_at')::timestamptz,
        NULLIF(p_record->>'trial_started_at', '')::timestamptz,
        NULLIF(p_record->>'cancelled_at', '')::timestamptz,
        p_record->>'cancellation_reason', p_operation_id, p_run_id,
        v_run.snapshot_at
    );
    v_sha := platform.billing_v1_adoption_tenant_fingerprint_v1(
        'subscription', v_id
    );
    PERFORM platform.billing_record_v1_adoption_mapping_v1(
        p_operation_id, p_run_id, 'subscriptions', p_record->>'source_id',
        'subscription', v_id, v_sha, v_run.snapshot_at
    );
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_apply_v1_subscription_v1(
    uuid, uuid, jsonb
) FROM PUBLIC;
-- +goose StatementEnd
