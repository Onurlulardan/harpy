-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_challenge_resolve_v1(
    p_scope_schema name,
    p_challenge_id uuid,
    p_expected_revision bigint,
    p_expected_attempt_count smallint,
    p_disposition text
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_state text;
    v_revision bigint;
    v_attempt_count smallint;
    v_expires_at timestamptz;
    v_next_attempt integer;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       NOT (
           p_disposition OPERATOR(pg_catalog.=) 'consume' OR
           p_disposition OPERATOR(pg_catalog.=) 'reject' OR
           p_disposition OPERATOR(pg_catalog.=) 'cancel' OR
           p_disposition OPERATOR(pg_catalog.=) 'expire'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM challenge resolution scope or disposition is invalid';
    END IF;
    SELECT challenge.state, challenge.revision,
           challenge.attempt_count, challenge.expires_at
    INTO v_state, v_revision, v_attempt_count, v_expires_at
    FROM iam_auth_challenges AS challenge
    WHERE challenge.id OPERATOR(pg_catalog.=) p_challenge_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM challenge is unknown';
    END IF;
    IF v_state IS DISTINCT FROM 'pending' OR
       v_revision IS DISTINCT FROM p_expected_revision OR
       v_attempt_count IS DISTINCT FROM p_expected_attempt_count THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM challenge resolution CAS is stale';
    END IF;
    IF v_now OPERATOR(pg_catalog.>=) v_expires_at THEN
        UPDATE iam_auth_challenges
        SET state = 'expired', terminal_at = v_now,
            purge_after = v_now OPERATOR(pg_catalog.+) INTERVAL '24 hours',
            updated_at = v_now,
            revision = v_revision OPERATOR(pg_catalog.+) 1
        WHERE id OPERATOR(pg_catalog.=) p_challenge_id;
        RETURN 'expired';
    END IF;
    IF p_disposition OPERATOR(pg_catalog.=) 'expire' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM challenge has not expired';
    END IF;
    IF p_disposition OPERATOR(pg_catalog.=) 'consume' THEN
        UPDATE iam_auth_challenges
        SET state = 'consumed', terminal_at = v_now, updated_at = v_now,
            revision = v_revision OPERATOR(pg_catalog.+) 1
        WHERE id OPERATOR(pg_catalog.=) p_challenge_id;
        RETURN 'consumed';
    END IF;
    IF p_disposition OPERATOR(pg_catalog.=) 'cancel' THEN
        UPDATE iam_auth_challenges
        SET state = 'cancelled', terminal_at = v_now, updated_at = v_now,
            revision = v_revision OPERATOR(pg_catalog.+) 1
        WHERE id OPERATOR(pg_catalog.=) p_challenge_id;
        RETURN 'cancelled';
    END IF;
    v_next_attempt := v_attempt_count OPERATOR(pg_catalog.+) 1;
    IF v_next_attempt OPERATOR(pg_catalog.>=) 5 THEN
        UPDATE iam_auth_challenges
        SET state = 'failed', attempt_count = 5,
            terminal_at = v_now, updated_at = v_now,
            revision = v_revision OPERATOR(pg_catalog.+) 1
        WHERE id OPERATOR(pg_catalog.=) p_challenge_id;
        RETURN 'failed';
    END IF;
    UPDATE iam_auth_challenges
    SET attempt_count = v_next_attempt, updated_at = v_now,
        revision = v_revision OPERATOR(pg_catalog.+) 1
    WHERE id OPERATOR(pg_catalog.=) p_challenge_id;
    RETURN 'pending';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_challenge_resolve_v1(
    name,uuid,bigint,smallint,text
) FROM PUBLIC;
