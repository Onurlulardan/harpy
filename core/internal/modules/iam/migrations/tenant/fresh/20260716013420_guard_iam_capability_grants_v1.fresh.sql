-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_guard_capability_grant_head_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    current_disposition text;
BEGIN
    IF TG_OP NOT IN ('INSERT', 'UPDATE') THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM capability grant head guard invocation is invalid';
    END IF;
    SELECT revision.disposition INTO current_disposition
    FROM iam_capability_grant_revisions AS revision
    WHERE revision.grant_id = NEW.id
      AND revision.grant_revision = NEW.current_revision
      AND revision.principal_id = NEW.principal_id
      AND revision.capability_key = NEW.capability_key;
    IF current_disposition IS NULL OR current_disposition IS DISTINCT FROM NEW.state THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_capability_grants_head_guard',
            MESSAGE = 'IAM capability grant head does not match its current revision';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_guard_capability_grant_head_v1() FROM PUBLIC;

CREATE TRIGGER iam_capability_grants_revision_guard
BEFORE UPDATE ON iam_capability_grants
FOR EACH ROW EXECUTE FUNCTION iam_guard_revision_step_v1();
ALTER TABLE iam_capability_grants
    ENABLE ALWAYS TRIGGER iam_capability_grants_revision_guard;

CREATE TRIGGER iam_capability_grants_immutable_guard
BEFORE UPDATE ON iam_capability_grants
FOR EACH ROW EXECUTE FUNCTION iam_guard_immutable_columns_v1(
    'id', 'principal_id', 'capability_key', 'created_at'
);
ALTER TABLE iam_capability_grants
    ENABLE ALWAYS TRIGGER iam_capability_grants_immutable_guard;

CREATE TRIGGER iam_capability_grants_state_guard
BEFORE UPDATE ON iam_capability_grants
FOR EACH ROW EXECUTE FUNCTION iam_guard_state_transition_v1('active>revoked');
ALTER TABLE iam_capability_grants
    ENABLE ALWAYS TRIGGER iam_capability_grants_state_guard;

CREATE CONSTRAINT TRIGGER iam_capability_grants_head_guard
AFTER INSERT OR UPDATE ON iam_capability_grants
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION iam_guard_capability_grant_head_v1();

CREATE TRIGGER iam_capability_grants_removal_guard
BEFORE DELETE OR TRUNCATE ON iam_capability_grants
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_capability_grants
    ENABLE ALWAYS TRIGGER iam_capability_grants_removal_guard;

CREATE TRIGGER iam_capability_grant_revisions_mutation_guard
BEFORE UPDATE OR DELETE OR TRUNCATE ON iam_capability_grant_revisions
FOR EACH STATEMENT EXECUTE FUNCTION iam_forbid_rbac_mutation_v1();
ALTER TABLE iam_capability_grant_revisions
    ENABLE ALWAYS TRIGGER iam_capability_grant_revisions_mutation_guard;
