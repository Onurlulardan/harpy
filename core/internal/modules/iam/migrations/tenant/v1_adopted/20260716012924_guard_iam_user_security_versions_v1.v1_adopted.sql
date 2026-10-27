-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_user_security_versions_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR
       TG_TABLE_NAME IS DISTINCT FROM 'iam_users' OR
       NEW.token_version < OLD.token_version OR
       NEW.authorization_version < OLD.authorization_version THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            CONSTRAINT = 'iam_users_security_version_guard',
            MESSAGE = 'IAM user security versions cannot decrease';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_user_security_versions_v1() FROM PUBLIC;

CREATE TRIGGER iam_users_security_version_guard
BEFORE UPDATE ON iam_users
FOR EACH ROW EXECUTE FUNCTION iam_guard_user_security_versions_v1();
ALTER TABLE iam_users
    ENABLE ALWAYS TRIGGER iam_users_security_version_guard;
