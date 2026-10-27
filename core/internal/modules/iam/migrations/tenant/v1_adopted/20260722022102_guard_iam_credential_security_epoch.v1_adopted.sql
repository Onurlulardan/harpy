-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_credential_security_epoch_guard_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_current bytea;
BEGIN
    v_current := iam_security_epoch_current_v1(pg_catalog.current_schema());
    IF TG_OP = 'INSERT' AND NEW.security_epoch IS DISTINCT FROM v_current THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM credential security epoch is stale';
    END IF;
    IF TG_OP = 'UPDATE' AND (
        NEW.security_epoch IS DISTINCT FROM OLD.security_epoch OR
        (NEW.state IN ('active', 'pending') AND
         NEW.security_epoch IS DISTINCT FROM v_current)
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM credential security epoch is stale';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER iam_sessions_security_epoch_guard
BEFORE INSERT OR UPDATE ON iam_sessions
FOR EACH ROW EXECUTE FUNCTION iam_credential_security_epoch_guard_v1();
CREATE TRIGGER iam_refresh_families_security_epoch_guard
BEFORE INSERT OR UPDATE ON iam_refresh_families
FOR EACH ROW EXECUTE FUNCTION iam_credential_security_epoch_guard_v1();
CREATE TRIGGER iam_refresh_tokens_security_epoch_guard
BEFORE INSERT OR UPDATE ON iam_refresh_tokens
FOR EACH ROW EXECUTE FUNCTION iam_credential_security_epoch_guard_v1();
CREATE TRIGGER iam_one_time_tokens_security_epoch_guard
BEFORE INSERT OR UPDATE ON iam_one_time_tokens
FOR EACH ROW EXECUTE FUNCTION iam_credential_security_epoch_guard_v1();
CREATE TRIGGER iam_auth_challenges_security_epoch_guard
BEFORE INSERT OR UPDATE ON iam_auth_challenges
FOR EACH ROW EXECUTE FUNCTION iam_credential_security_epoch_guard_v1();

ALTER TABLE iam_sessions ENABLE ALWAYS TRIGGER iam_sessions_security_epoch_guard;
ALTER TABLE iam_refresh_families
    ENABLE ALWAYS TRIGGER iam_refresh_families_security_epoch_guard;
ALTER TABLE iam_refresh_tokens
    ENABLE ALWAYS TRIGGER iam_refresh_tokens_security_epoch_guard;
ALTER TABLE iam_one_time_tokens
    ENABLE ALWAYS TRIGGER iam_one_time_tokens_security_epoch_guard;
ALTER TABLE iam_auth_challenges
    ENABLE ALWAYS TRIGGER iam_auth_challenges_security_epoch_guard;
REVOKE ALL ON FUNCTION iam_credential_security_epoch_guard_v1() FROM PUBLIC;
