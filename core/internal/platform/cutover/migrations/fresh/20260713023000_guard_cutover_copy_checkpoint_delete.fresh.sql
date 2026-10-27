-- +goose Up
CREATE TRIGGER cutover_copy_checkpoints_delete_guard
BEFORE DELETE ON platform.cutover_copy_checkpoints
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_delete();
