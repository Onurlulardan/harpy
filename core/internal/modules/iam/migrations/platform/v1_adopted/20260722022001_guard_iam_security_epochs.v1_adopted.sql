-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_security_epoch_guard_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM security epoch removal is forbidden';
    END IF;
    IF TG_OP = 'INSERT' AND (
        NEW.generation IS DISTINCT FROM 1 OR
        NEW.rotated_at IS DISTINCT FROM NEW.created_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM initial security epoch is invalid';
    END IF;
    IF TG_OP = 'UPDATE' AND (
        NEW.environment_id IS DISTINCT FROM OLD.environment_id OR
        NEW.created_at IS DISTINCT FROM OLD.created_at OR
        NEW.generation IS DISTINCT FROM OLD.generation + 1 OR
        NEW.security_epoch IS NOT DISTINCT FROM OLD.security_epoch OR
        NEW.rotated_at <= OLD.rotated_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM security epoch rotation is invalid';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER iam_security_epochs_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.iam_security_epochs
FOR EACH ROW EXECUTE FUNCTION platform.iam_security_epoch_guard_v1();
ALTER TABLE platform.iam_security_epochs
    ENABLE ALWAYS TRIGGER iam_security_epochs_guard;
REVOKE ALL ON FUNCTION platform.iam_security_epoch_guard_v1() FROM PUBLIC;
