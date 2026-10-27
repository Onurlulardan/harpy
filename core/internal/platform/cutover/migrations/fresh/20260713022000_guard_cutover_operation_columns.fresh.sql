-- +goose Up
CREATE TRIGGER cutover_operation_columns_immutable_guard
BEFORE UPDATE OR DELETE ON platform.cutover_operation_columns
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_mutation();
