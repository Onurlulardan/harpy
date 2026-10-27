-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_v1_actor_bridge_reconcile_v1(
    p_expected_revision bigint
)
RETURNS TABLE (remaining_scope_count bigint, revision bigint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_retirement platform.contract_retirements%ROWTYPE;
    v_remaining bigint;
    v_revision bigint;
    v_now timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_migrator' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_expected_revision IS NULL OR p_expected_revision < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM V1 actor bridge reconcile input is invalid';
    END IF;
    PERFORM authority.scope_id, authority.domain
    FROM platform.domain_cutovers AS authority
    JOIN platform.tenants AS tenant
      ON tenant.id = authority.scope_id AND tenant.deleted_at IS NULL
    WHERE authority.scope_kind = 'tenant' AND authority.domain IN (
        'analytics','billing','catalog','documents','forms',
        'notifications','work','workflow'
    )
    ORDER BY authority.scope_id, authority.domain COLLATE "C"
    FOR SHARE OF authority;

    SELECT pg_catalog.count(*) INTO v_remaining
    FROM platform.domain_cutovers AS authority
    JOIN platform.tenants AS tenant
      ON tenant.id = authority.scope_id AND tenant.deleted_at IS NULL
    WHERE authority.scope_kind = 'tenant' AND authority.domain IN (
        'analytics','billing','catalog','documents','forms',
        'notifications','work','workflow'
    ) AND authority.state IN ('v1','shadow') AND authority.writer = 'v1';

    SELECT retirement.* INTO v_retirement
    FROM platform.contract_retirements AS retirement
    WHERE retirement.contract_key = 'iam.v1_actor_assertion.v1'
    FOR UPDATE;
    IF NOT FOUND OR v_retirement.owner_module IS DISTINCT FROM 'iam' OR
       v_retirement.domain IS DISTINCT FROM 'iam' OR
       v_retirement.state IS DISTINCT FROM 'active' OR
       v_retirement.revision IS DISTINCT FROM p_expected_revision THEN
        RAISE EXCEPTION USING ERRCODE = 'P5505',
            MESSAGE = 'IAM V1 actor bridge reconcile authority differs';
    END IF;
    IF EXISTS (
        SELECT 1 FROM platform.iam_v1_actor_assertion_receipts
        WHERE accounted_at IS NULL
    ) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5606',
            MESSAGE = 'IAM V1 actor bridge hits are not reconciled';
    END IF;
    IF v_retirement.revision = 9223372036854775807 THEN
        RAISE EXCEPTION USING ERRCODE = 'P5504',
            MESSAGE = 'IAM V1 actor bridge counter overflow';
    END IF;

    v_now := pg_catalog.clock_timestamp();
    UPDATE platform.contract_retirements AS retirement
    SET remaining_scope_count = v_remaining,
        zero_hit_since = CASE WHEN v_remaining = 0
            THEN COALESCE(retirement.zero_hit_since, v_now)
            ELSE NULL END,
        eligible_at = NULL,
        revision = retirement.revision + 1,
        updated_at = v_now
    WHERE retirement.contract_key = 'iam.v1_actor_assertion.v1'
    RETURNING retirement.revision INTO STRICT v_revision;
    RETURN QUERY SELECT v_remaining, v_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_v1_actor_bridge_reconcile_v1(bigint)
FROM PUBLIC;
