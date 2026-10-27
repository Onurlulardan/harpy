-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_v1_actor_bridge_mark_eligible_v1(
    p_expected_revision bigint
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_retirement platform.contract_retirements%ROWTYPE;
    v_revision bigint;
    v_now timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_migrator' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_expected_revision IS NULL OR p_expected_revision < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM V1 actor bridge eligibility input is invalid';
    END IF;
    SELECT retirement.* INTO v_retirement
    FROM platform.contract_retirements AS retirement
    WHERE retirement.contract_key = 'iam.v1_actor_assertion.v1'
    FOR UPDATE;
    v_now := pg_catalog.clock_timestamp();
    IF NOT FOUND OR v_retirement.owner_module IS DISTINCT FROM 'iam' OR
       v_retirement.domain IS DISTINCT FROM 'iam' OR
       v_retirement.state IS DISTINCT FROM 'active' OR
       v_retirement.revision IS DISTINCT FROM p_expected_revision OR
       v_retirement.remaining_scope_count IS DISTINCT FROM 0 OR
       v_retirement.zero_hit_since IS NULL OR
       v_retirement.zero_hit_since + INTERVAL '30 days' > v_now OR
       EXISTS (
           SELECT 1 FROM platform.iam_v1_actor_assertion_receipts AS receipt
           WHERE receipt.accounted_at IS NULL OR
                 receipt.consumed_at >= v_retirement.zero_hit_since
       ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5607',
            MESSAGE = 'IAM V1 actor bridge is not eligible';
    END IF;
    IF v_retirement.revision = 9223372036854775807 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5504',
            MESSAGE = 'IAM V1 actor bridge counter overflow';
    END IF;
    UPDATE platform.contract_retirements AS retirement
    SET state = 'eligible', eligible_at = v_now,
        revision = retirement.revision + 1, updated_at = v_now
    WHERE retirement.contract_key = 'iam.v1_actor_assertion.v1'
    RETURNING retirement.revision INTO STRICT v_revision;
    RETURN v_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.iam_v1_actor_bridge_mark_eligible_v1(bigint)
FROM PUBLIC;

