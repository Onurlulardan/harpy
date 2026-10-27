-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_reauthentication_reject_v1(
    p_scope_schema name, p_challenge_id uuid,
    p_expected_revision bigint, p_expected_attempt_count smallint,
    p_authenticator_id uuid, p_expected_authenticator_revision bigint,
    p_quarantine_authenticator boolean, p_expire boolean
)
RETURNS TABLE (state text, revision bigint, attempt_count smallint)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_authenticator_id uuid;
    v_authenticator_state text;
    v_authenticator_revision bigint;
    v_disposition text;
    v_resolved text;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_challenge_id IS NULL OR p_expected_revision OPERATOR(pg_catalog.<=) 0 OR
       p_expected_attempt_count NOT BETWEEN 0 AND 4 OR
       p_quarantine_authenticator IS NULL OR p_expire IS NULL OR
       (p_quarantine_authenticator AND (
           p_authenticator_id IS NULL OR
           p_expected_authenticator_revision OPERATOR(pg_catalog.<=) 0
       )) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM reauthentication rejection evidence is invalid';
    END IF;
    SELECT challenge.authenticator_id
    INTO v_authenticator_id
    FROM iam_auth_challenges AS challenge
    WHERE challenge.id OPERATOR(pg_catalog.=) p_challenge_id
      AND challenge.purpose OPERATOR(pg_catalog.=) 'reauth_v1'
    FOR UPDATE;
    IF NOT FOUND OR (p_quarantine_authenticator AND
        v_authenticator_id IS DISTINCT FROM p_authenticator_id) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM reauthentication rejection CAS is stale';
    END IF;
    IF p_quarantine_authenticator THEN
        SELECT authenticator.state, authenticator.revision
        INTO v_authenticator_state, v_authenticator_revision
        FROM iam_authenticators AS authenticator
        WHERE authenticator.id OPERATOR(pg_catalog.=) p_authenticator_id
        FOR UPDATE;
    END IF;
    IF p_quarantine_authenticator AND (
        v_authenticator_state IS DISTINCT FROM 'active' OR
        v_authenticator_revision IS DISTINCT FROM
            p_expected_authenticator_revision
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM reauthentication rejection CAS is stale';
    END IF;
    v_disposition := CASE WHEN p_expire THEN 'expire' ELSE 'reject' END;
    v_resolved := iam_challenge_resolve_v1(
        p_scope_schema, p_challenge_id, p_expected_revision,
        p_expected_attempt_count, v_disposition
    );
    IF p_quarantine_authenticator THEN
        UPDATE iam_authenticators
        SET state = 'compromised', compromised_at = v_now,
            updated_at = v_now,
            revision = p_expected_authenticator_revision
                OPERATOR(pg_catalog.+) 1
        WHERE id OPERATOR(pg_catalog.=) p_authenticator_id
          AND state OPERATOR(pg_catalog.=) 'active'
          AND revision OPERATOR(pg_catalog.=)
              p_expected_authenticator_revision;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM authenticator quarantine CAS is stale';
        END IF;
    END IF;
    RETURN QUERY
    SELECT challenge.state, challenge.revision, challenge.attempt_count
    FROM iam_auth_challenges AS challenge
    WHERE challenge.id OPERATOR(pg_catalog.=) p_challenge_id
      AND challenge.state OPERATOR(pg_catalog.=) v_resolved;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_reauthentication_reject_v1(
    name,uuid,bigint,smallint,uuid,bigint,boolean,boolean
) FROM PUBLIC;
