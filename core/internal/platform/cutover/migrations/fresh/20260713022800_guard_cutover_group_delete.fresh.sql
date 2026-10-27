-- +goose Up
CREATE TRIGGER cutover_groups_delete_guard
BEFORE DELETE ON platform.cutover_groups
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_delete();
