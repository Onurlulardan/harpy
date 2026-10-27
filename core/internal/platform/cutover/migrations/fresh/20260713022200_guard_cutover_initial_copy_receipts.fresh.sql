-- +goose Up
CREATE TRIGGER cutover_initial_copy_receipts_immutable_guard
BEFORE UPDATE OR DELETE ON platform.cutover_initial_copy_receipts
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_mutation();
