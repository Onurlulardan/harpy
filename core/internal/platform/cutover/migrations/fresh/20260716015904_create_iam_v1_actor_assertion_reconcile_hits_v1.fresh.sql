-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_v1_actor_assertion_reconcile_hits_v1(
    p_limit integer
)
RETURNS TABLE (accounted_count bigint, latest_hit_at timestamptz)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_retirement platform.contract_retirements%ROWTYPE;
    v_count bigint;
    v_latest timestamptz;
    v_now timestamptz;
    v_receipt_tenant_id uuid;
    v_receipt_jti uuid;
    v_receipt_consumed_at timestamptz;
BEGIN
    IF SESSION_USER NOT IN ('forgecrud_v2_runtime','forgecrud_migrator') OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_limit IS NULL OR p_limit < 1 OR p_limit > 10000 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM V1 actor assertion hit reconcile input is invalid';
    END IF;
    SELECT retirement.* INTO v_retirement
    FROM platform.contract_retirements AS retirement
    WHERE retirement.contract_key = 'iam.v1_actor_assertion.v1'
    FOR UPDATE;
    IF NOT FOUND OR v_retirement.owner_module IS DISTINCT FROM 'iam' OR
       v_retirement.domain IS DISTINCT FROM 'iam' OR
       v_retirement.state IS DISTINCT FROM 'active' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5603',
            MESSAGE = 'IAM V1 actor bridge contract is unavailable';
    END IF;

    v_count := 0;
    v_latest := NULL;
    FOR v_receipt_tenant_id, v_receipt_jti, v_receipt_consumed_at IN
        SELECT receipt.tenant_id, receipt.jti, receipt.consumed_at
        FROM platform.iam_v1_actor_assertion_receipts AS receipt
        WHERE receipt.accounted_at IS NULL
        ORDER BY receipt.consumed_at, receipt.tenant_id, receipt.jti
        LIMIT p_limit
        FOR UPDATE
    LOOP
        IF v_retirement.hit_count = 9223372036854775807 OR
           v_retirement.revision = 9223372036854775807 THEN
            RAISE EXCEPTION USING ERRCODE = 'P5504',
                MESSAGE = 'IAM V1 actor bridge counter overflow';
        END IF;
        v_now := pg_catalog.clock_timestamp();
        UPDATE platform.iam_v1_actor_assertion_receipts AS receipt
        SET accounted_at = v_now
        WHERE receipt.tenant_id = v_receipt_tenant_id
          AND receipt.jti = v_receipt_jti;
        UPDATE platform.contract_retirements AS retirement
        SET hit_count = retirement.hit_count + 1,
            last_hit_at = v_now,
            revision = retirement.revision + 1,
            updated_at = v_now
        WHERE retirement.contract_key = 'iam.v1_actor_assertion.v1';
        v_retirement.hit_count := v_retirement.hit_count + 1;
        v_retirement.revision := v_retirement.revision + 1;
        v_count := v_count + 1;
        v_latest := CASE
            WHEN v_latest IS NULL OR v_latest < v_receipt_consumed_at
            THEN v_receipt_consumed_at ELSE v_latest
        END;
    END LOOP;
    RETURN QUERY SELECT v_count, v_latest;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.iam_v1_actor_assertion_reconcile_hits_v1(integer)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.iam_v1_actor_assertion_reconcile_hits_v1(integer)
TO forgecrud_runtime;
