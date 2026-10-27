-- +goose Up
CREATE CONSTRAINT TRIGGER cutover_groups_transition_event_guard
AFTER UPDATE ON platform.cutover_groups
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_require_group_transition_event();
