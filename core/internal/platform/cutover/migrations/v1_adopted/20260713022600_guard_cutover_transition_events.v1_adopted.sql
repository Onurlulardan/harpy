-- +goose Up
CREATE TRIGGER cutover_transition_events_immutable_guard
BEFORE UPDATE OR DELETE ON platform.cutover_transition_events
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_mutation();
