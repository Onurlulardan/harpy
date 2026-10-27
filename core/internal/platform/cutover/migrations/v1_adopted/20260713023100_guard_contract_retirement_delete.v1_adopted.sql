-- +goose Up
CREATE TRIGGER contract_retirements_delete_guard
BEFORE DELETE ON platform.contract_retirements
FOR EACH ROW
EXECUTE FUNCTION platform.cutover_reject_delete();
