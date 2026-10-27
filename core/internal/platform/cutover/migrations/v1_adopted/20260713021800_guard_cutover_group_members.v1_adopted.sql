-- +goose Up
CREATE TRIGGER cutover_group_members_immutable_guard
BEFORE UPDATE OR DELETE ON platform.cutover_group_members
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_mutation();
