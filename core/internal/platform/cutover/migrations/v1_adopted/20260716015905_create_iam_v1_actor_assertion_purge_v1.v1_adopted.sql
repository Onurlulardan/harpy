-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_v1_actor_assertion_purge_v1(
    p_limit integer
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_deleted bigint;
    v_cutoff timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_limit IS NULL OR p_limit < 1 OR p_limit > 10000 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM V1 actor assertion purge input is invalid';
    END IF;
    v_cutoff := pg_catalog.clock_timestamp() - INTERVAL '24 hours';
    WITH candidates AS MATERIALIZED (
        SELECT receipt.tenant_id, receipt.jti
        FROM platform.iam_v1_actor_assertion_receipts AS receipt
        WHERE receipt.accounted_at IS NOT NULL
          AND receipt.expires_at < v_cutoff
        ORDER BY receipt.expires_at, receipt.tenant_id, receipt.jti
        LIMIT p_limit
        FOR UPDATE SKIP LOCKED
    )
    DELETE FROM platform.iam_v1_actor_assertion_receipts AS receipt
    USING candidates
    WHERE receipt.tenant_id = candidates.tenant_id
      AND receipt.jti = candidates.jti;
    GET DIAGNOSTICS v_deleted = ROW_COUNT;
    RETURN v_deleted;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_v1_actor_assertion_purge_v1(integer)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_v1_actor_assertion_purge_v1(integer)
TO harpy_runtime;

