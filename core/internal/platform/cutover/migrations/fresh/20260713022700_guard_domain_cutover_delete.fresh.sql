-- +goose Up
CREATE TRIGGER domain_cutovers_delete_guard
BEFORE DELETE ON platform.domain_cutovers
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_delete();
