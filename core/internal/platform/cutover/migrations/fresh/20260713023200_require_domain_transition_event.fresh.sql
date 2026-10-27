-- +goose Up
CREATE CONSTRAINT TRIGGER domain_cutovers_transition_event_guard
AFTER UPDATE ON platform.domain_cutovers
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_require_domain_transition_event();
