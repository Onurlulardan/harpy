-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_worker_capability_commit_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM iam_profile_backfill_worker_capabilities AS capability
        WHERE capability.backend_pid = NEW.backend_pid
          AND capability.transaction_id = NEW.transaction_id
          AND capability.plan_id = NEW.plan_id
          AND capability.resource_id = NEW.resource_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile worker capability was not consumed';
    END IF;
    RETURN NULL;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_worker_capability_commit_v1()
    FROM PUBLIC;

CREATE CONSTRAINT TRIGGER iam_profile_worker_capability_commit_guard
AFTER INSERT ON iam_profile_backfill_worker_capabilities
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION iam_profile_worker_capability_commit_v1();
ALTER TABLE iam_profile_backfill_worker_capabilities ENABLE ALWAYS TRIGGER
    iam_profile_worker_capability_commit_guard;
