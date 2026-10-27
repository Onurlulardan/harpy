-- +goose Up
CREATE TRIGGER applied_change_receipts_immutable_guard
BEFORE UPDATE OR DELETE ON platform.applied_change_receipts
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_mutation();
