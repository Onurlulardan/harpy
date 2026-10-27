-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_state_transition_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    transition text;
    transition_index integer;
    allowed boolean := false;
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR TG_NARGS < 1 THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000',
            MESSAGE = 'IAM state transition guard invocation is invalid';
    END IF;
    IF NEW.state IS NOT DISTINCT FROM OLD.state THEN
        RETURN NEW;
    END IF;
    transition := OLD.state || '>' || NEW.state;
    FOR transition_index IN 0..TG_NARGS - 1 LOOP
        IF transition IS NOT DISTINCT FROM TG_ARGV[transition_index] THEN
            allowed := true;
            EXIT;
        END IF;
    END LOOP;
    IF NOT allowed THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            CONSTRAINT = TG_TABLE_NAME || '_state_transition_guard',
            MESSAGE = pg_catalog.format(
                'IAM state transition %s is forbidden',
                transition
            );
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_state_transition_v1() FROM PUBLIC;

CREATE TRIGGER iam_sessions_state_transition_guard
BEFORE UPDATE ON iam_sessions
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1(
    'active>revoked', 'active>expired'
);
ALTER TABLE iam_sessions
    ENABLE ALWAYS TRIGGER iam_sessions_state_transition_guard;

CREATE TRIGGER iam_refresh_families_state_transition_guard
BEFORE UPDATE ON iam_refresh_families
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1(
    'active>revoked', 'active>compromised', 'active>expired'
);
ALTER TABLE iam_refresh_families
    ENABLE ALWAYS TRIGGER iam_refresh_families_state_transition_guard;

CREATE TRIGGER iam_refresh_tokens_state_transition_guard
BEFORE UPDATE ON iam_refresh_tokens
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1(
    'active>consumed', 'active>revoked', 'active>expired'
);
ALTER TABLE iam_refresh_tokens
    ENABLE ALWAYS TRIGGER iam_refresh_tokens_state_transition_guard;

CREATE TRIGGER iam_recovery_codes_state_transition_guard
BEFORE UPDATE ON iam_recovery_codes
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1(
    'active>consumed', 'active>revoked'
);
ALTER TABLE iam_recovery_codes
    ENABLE ALWAYS TRIGGER iam_recovery_codes_state_transition_guard;

CREATE TRIGGER iam_auth_challenges_state_transition_guard
BEFORE UPDATE ON iam_auth_challenges
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1(
    'pending>consumed', 'pending>failed', 'pending>expired',
    'pending>cancelled'
);
ALTER TABLE iam_auth_challenges
    ENABLE ALWAYS TRIGGER iam_auth_challenges_state_transition_guard;

CREATE TRIGGER iam_one_time_tokens_state_transition_guard
BEFORE UPDATE ON iam_one_time_tokens
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1(
    'pending>consumed', 'pending>revoked', 'pending>expired'
);
ALTER TABLE iam_one_time_tokens
    ENABLE ALWAYS TRIGGER iam_one_time_tokens_state_transition_guard;

CREATE TRIGGER iam_invitations_state_transition_guard
BEFORE UPDATE ON iam_invitations
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1(
    'pending>accepted', 'pending>revoked', 'pending>expired'
);
ALTER TABLE iam_invitations
    ENABLE ALWAYS TRIGGER iam_invitations_state_transition_guard;
