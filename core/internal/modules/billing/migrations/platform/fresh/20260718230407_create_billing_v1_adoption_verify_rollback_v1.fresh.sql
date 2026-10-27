-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_v1_adoption_verify_rollback_v1(
    p_operation_id uuid,
    p_run_id uuid,
    p_source_sha256 text
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_run platform.billing_v1_adoption_runs%ROWTYPE;
    v_source_count bigint;
    v_target_count bigint;
    v_source_sha text;
    v_target_sha text;
    v_money bigint;
    v_period bigint;
    v_coupon bigint;
    v_usage bigint;
    v_receipt_id uuid;
    v_now timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_source_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing V1 adoption verification principal denied';
    END IF;
    SELECT run.* INTO STRICT v_run
    FROM platform.billing_v1_adoption_runs AS run
    WHERE run.operation_id = p_operation_id AND run.run_id = p_run_id
    FOR UPDATE;
    IF v_run.state = 'rollback_rehearsed' THEN
        IF v_run.source_sha256 IS DISTINCT FROM
               pg_catalog.decode(p_source_sha256, 'hex') OR
           v_run.baseline_sha256 IS DISTINCT FROM v_run.target_sha256 OR
           v_run.money_difference <> 0 OR v_run.period_difference <> 0 OR
           v_run.coupon_difference <> 0 OR v_run.usage_difference <> 0 THEN
            RAISE EXCEPTION USING ERRCODE = '22000',
                MESSAGE = 'billing V1 adoption completed evidence differs';
        END IF;
        RETURN pg_catalog.encode(v_run.target_sha256, 'hex');
    END IF;
    SELECT * INTO STRICT v_source_count, v_target_count,
        v_source_sha, v_target_sha, v_money, v_period, v_coupon, v_usage
    FROM platform.billing_verify_v1_adoption_v1(p_operation_id, p_run_id);
    IF v_source_count IS DISTINCT FROM v_target_count OR
       v_source_sha IS DISTINCT FROM v_target_sha OR
       v_money <> 0 OR v_period <> 0 OR v_coupon <> 0 OR v_usage <> 0 THEN
        RAISE EXCEPTION USING ERRCODE = '22000',
            MESSAGE = 'billing V1 adoption reconciliation is not zero-diff',
            DETAIL = pg_catalog.format(
                'source=%s target=%s digest_equal=%s money=%s period=%s coupon=%s usage=%s',
                v_source_count, v_target_count, v_source_sha = v_target_sha,
                v_money, v_period, v_coupon, v_usage
            );
    END IF;
    v_receipt_id := pg_catalog.gen_random_uuid();
    v_now := pg_catalog.clock_timestamp();
    PERFORM * FROM platform.billing_rehearse_v1_adoption_rollback_v1(
        p_operation_id, p_run_id, v_receipt_id, p_source_sha256, v_now
    );
    PERFORM platform.billing_assert_v1_adoption_ready_v1(
        p_operation_id, p_run_id, p_source_sha256
    );
    RETURN v_target_sha;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_v1_adoption_verify_rollback_v1(
    uuid, uuid, text
) FROM PUBLIC;
