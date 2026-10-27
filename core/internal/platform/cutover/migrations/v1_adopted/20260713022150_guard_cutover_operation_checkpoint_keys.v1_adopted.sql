-- +goose Up
CREATE TRIGGER cutover_operation_checkpoint_keys_immutable_guard
BEFORE UPDATE OR DELETE ON platform.cutover_operation_checkpoint_keys
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_mutation();
