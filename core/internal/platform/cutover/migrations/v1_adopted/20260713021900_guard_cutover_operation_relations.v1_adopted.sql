-- +goose Up
CREATE TRIGGER cutover_operation_relations_immutable_guard
BEFORE UPDATE OR DELETE ON platform.cutover_operation_relations
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_mutation();
