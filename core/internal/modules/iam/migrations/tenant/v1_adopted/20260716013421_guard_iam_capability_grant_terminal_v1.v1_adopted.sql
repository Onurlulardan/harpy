-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_capability_grant_terminal_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR OLD.state IS DISTINCT FROM 'revoked' THEN
        RETURN NEW;
    END IF;
    RAISE EXCEPTION USING ERRCODE = '23514',
        CONSTRAINT = 'iam_capability_grants_terminal_guard',
        MESSAGE = 'Revoked IAM capability grant is immutable';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_capability_grant_terminal_v1() FROM PUBLIC;

CREATE TRIGGER iam_capability_grants_terminal_guard
BEFORE UPDATE ON iam_capability_grants
FOR EACH ROW EXECUTE FUNCTION iam_guard_capability_grant_terminal_v1();
ALTER TABLE iam_capability_grants
    ENABLE ALWAYS TRIGGER iam_capability_grants_terminal_guard;
