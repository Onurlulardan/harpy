-- +goose Up
CREATE TRIGGER cutover_operations_delete_guard
BEFORE DELETE ON platform.cutover_operations
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_delete();
