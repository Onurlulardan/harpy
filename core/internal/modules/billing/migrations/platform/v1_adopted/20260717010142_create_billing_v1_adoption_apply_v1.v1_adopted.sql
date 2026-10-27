-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_apply_v1_adoption_v1(
    p_operation_id uuid,
    p_run_id uuid
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_record jsonb;
    v_count bigint;
    v_now timestamptz;
BEGIN
    v_now := pg_catalog.clock_timestamp();
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
    FOR UPDATE;
    IF v_run.state IN ('applied', 'verified', 'rollback_rehearsed') THEN
        RETURN v_run.applied_count;
    END IF;
    IF v_run.state <> 'planned' OR NOT v_run.eligible THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'billing V1 adoption run is not applicable';
    END IF;
    IF v_run.scope_kind = 'global' THEN
        PERFORM platform.billing_apply_v1_provider_account_v1(
            p_operation_id, p_run_id,
            pg_catalog.to_jsonb(v_run.plan_document->>'manual_provider_account_id')
        );
        FOR v_record IN
            SELECT value FROM pg_catalog.jsonb_array_elements(
                v_run.plan_document->'offerings'
            )
        LOOP
            PERFORM platform.billing_apply_v1_offering_v1(
                p_operation_id, p_run_id, v_record
            );
        END LOOP;
        FOR v_record IN
            SELECT value FROM pg_catalog.jsonb_array_elements(
                v_run.plan_document->'discounts'
            )
        LOOP
            PERFORM platform.billing_apply_v1_discount_v1(
                p_operation_id, p_run_id, v_record
            );
        END LOOP;
        FOR v_record IN
            SELECT value FROM pg_catalog.jsonb_array_elements(
                v_run.plan_document->'campaigns'
            )
        LOOP
            PERFORM platform.billing_apply_v1_campaign_v1(
                p_operation_id, p_run_id, v_record
            );
        END LOOP;
    ELSE
        FOR v_record IN
            SELECT value FROM pg_catalog.jsonb_array_elements(
                v_run.plan_document->'base_assignments'
            )
        LOOP
            PERFORM platform.billing_apply_v1_base_assignment_v1(
                p_operation_id, p_run_id, v_record
            );
        END LOOP;
        FOR v_record IN
            SELECT value FROM pg_catalog.jsonb_array_elements(
                v_run.plan_document->'subscriptions'
            )
        LOOP
            PERFORM platform.billing_apply_v1_subscription_v1(
                p_operation_id, p_run_id, v_record
            );
        END LOOP;
        FOR v_record IN
            SELECT value FROM pg_catalog.jsonb_array_elements(
                v_run.plan_document->'redemptions'
            )
        LOOP
            PERFORM platform.billing_apply_v1_redemption_v1(
                p_operation_id, p_run_id, v_record
            );
        END LOOP;
        FOR v_record IN
            SELECT value FROM pg_catalog.jsonb_array_elements(
                v_run.plan_document->'usage_openings'
            )
        LOOP
            PERFORM platform.billing_apply_v1_opening_v1(
                p_operation_id, p_run_id, 'usage_opening', v_record
            );
        END LOOP;
        FOR v_record IN
            SELECT value FROM pg_catalog.jsonb_array_elements(
                v_run.plan_document->'gauge_openings'
            )
        LOOP
            PERFORM platform.billing_apply_v1_opening_v1(
                p_operation_id, p_run_id, 'gauge_opening', v_record
            );
        END LOOP;
        FOR v_record IN
            SELECT value FROM pg_catalog.jsonb_array_elements(
                v_run.plan_document->'legacy_events'
            )
        LOOP
            PERFORM platform.billing_apply_v1_event_v1(
                p_operation_id, p_run_id, v_record
            );
        END LOOP;
    END IF;
    SELECT pg_catalog.count(*) INTO v_count
    FROM platform.billing_v1_adoption_mappings AS mapping
    WHERE mapping.operation_id = p_operation_id AND mapping.run_id = p_run_id;
    UPDATE platform.billing_v1_adoption_runs SET
        state = 'applied', applied_count = v_count,
        applied_at = v_now, updated_at = v_now
    WHERE operation_id = p_operation_id AND run_id = p_run_id;
    RETURN v_count;
END;
$function$;
REVOKE ALL ON FUNCTION platform.billing_apply_v1_adoption_v1(uuid, uuid)
    FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.billing_apply_v1_adoption_v1(uuid, uuid)
    TO harpy_adopter;
-- +goose StatementEnd
