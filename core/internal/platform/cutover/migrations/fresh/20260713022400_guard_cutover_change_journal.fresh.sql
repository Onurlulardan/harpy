-- +goose Up
CREATE TRIGGER cutover_change_journal_immutable_guard
BEFORE UPDATE OR DELETE ON platform.cutover_change_journal
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_mutation();
