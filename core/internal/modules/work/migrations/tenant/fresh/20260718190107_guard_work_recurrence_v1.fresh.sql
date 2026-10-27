-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION work_guard_recurrence_revision_v1() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog AS $$
BEGIN
    RAISE EXCEPTION USING ERRCODE = '42501', MESSAGE = 'WORK_RECURRENCE_REVISION_IMMUTABLE';
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER work_recurrence_revision_guard
BEFORE UPDATE OR DELETE ON work_recurrence_rule_revisions
FOR EACH ROW EXECUTE FUNCTION work_guard_recurrence_revision_v1();
ALTER TABLE work_recurrence_rule_revisions ENABLE ALWAYS TRIGGER work_recurrence_revision_guard;
